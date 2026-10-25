package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Sush1sui/FNS_BOT/internal/bot"
	"github.com/Sush1sui/FNS_BOT/internal/db"
	"github.com/Sush1sui/FNS_BOT/internal/utils"
	"github.com/jackc/pgx/v5"
)

const (
	authCookieName  = "fns_auth"
	stateCookieName = "fns_oauth_state"
	csrfCookieName  = "fns_csrf"
	csrfHeaderName  = "X-CSRF-Token"
	sessionTTL      = 7 * 24 * time.Hour
	lastSeenGap     = 30 * time.Minute
)

type AuthClaims struct {
	UserID      string `json:"uid"`
	Name        string `json:"name"`
	Email       string `json:"email,omitempty"`
	Image       string `json:"image,omitempty"`
	DiscordID   string `json:"discord_id"`
	AccessToken string `json:"-"`
}

func (c *AuthClaims) GetDiscordID() string   { return c.DiscordID }
func (c *AuthClaims) GetAccessToken() string { return c.AccessToken }

type (
	discordTokenResponse = utils.DiscordTokenResponse
	discordUser          = utils.DiscordUser
	discordGuild         = utils.DiscordGuild
	discordGuildMember   = utils.DiscordGuildMember
	serverSummary        = utils.ServerSummary
)

func (s *Server) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	// Start OAuth login and set CSRF state cookie.
	state, err := utils.RandomState(24)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to init auth"})
		return
	}

	utils.SetStateCookie(w, state, s.Config.CookieSecure, utils.CookieSameSite(s.Config.CookieSecure), stateCookieName, 300)

	authURL := fmt.Sprintf(
		"https://discord.com/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		url.QueryEscape(s.Config.DiscordClientID),
		url.QueryEscape(s.Config.DiscordRedirectURL),
		url.QueryEscape("identify email guilds"),
		url.QueryEscape(state),
	)

	http.Redirect(w, r, authURL, http.StatusFound)
}

func (s *Server) handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	// Exchange code for token, fetch user, mint session cookie.
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing code or state"})
		return
	}

	if !utils.ValidateStateCookie(r, state, stateCookieName) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid state"})
		return
	}

	token, err := utils.ExchangeDiscordToken(r.Context(), s.Config.DiscordClientID, s.Config.DiscordClientSecret, s.Config.DiscordRedirectURL, code)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "token exchange failed"})
		return
	}

	user, err := utils.FetchDiscordUser(r.Context(), token.AccessToken)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to fetch user"})
		return
	}

	name := user.GlobalName
	if name == "" {
		name = user.Username
	}

	image := ""
	if user.Avatar != "" {
		image = fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", user.ID, user.Avatar)
	}

	sessionID, err := utils.RandomState(32)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to init session"})
		return
	}

	encryptedToken, err := utils.EncryptTokenV1(s.Config.AccessTokenKey, token.AccessToken)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to encrypt session"})
		return
	}

	now := time.Now()
	if err := s.DB.CreateAuthSession(r.Context(), db.CreateAuthSessionParams{
		SessionID:   sessionID,
		UserID:      user.ID,
		DiscordID:   user.ID,
		Name:        name,
		Email:       user.Email,
		Image:       image,
		AccessToken: encryptedToken,
		CreatedAt:   now.Unix(),
		ExpiresAt:   now.Add(sessionTTL).Unix(),
		LastSeen:    now.Unix(),
	}); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create session"})
		return
	}

	utils.SetAuthCookie(w, sessionID, s.Config.CookieSecure, utils.CookieSameSite(s.Config.CookieSecure), authCookieName)
	csrfToken, err := utils.RandomState(32)
	if err == nil {
		utils.SetCSRFCookie(w, csrfToken, s.Config.CookieSecure, utils.CookieSameSite(s.Config.CookieSecure), csrfCookieName, int(sessionTTL.Seconds()))
	}
	utils.ClearStateCookie(w, s.Config.CookieSecure, utils.CookieSameSite(s.Config.CookieSecure), stateCookieName)

	http.Redirect(w, r, s.Config.ClientOrigin, http.StatusFound)
}

func (s *Server) handleAuthLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(authCookieName); err == nil && cookie.Value != "" {
		_ = s.DB.DeleteAuthSession(r.Context(), cookie.Value)
	}
	utils.ClearAuthCookie(w, s.Config.CookieSecure, utils.CookieSameSite(s.Config.CookieSecure), authCookieName)
	utils.ClearCSRFCookie(w, s.Config.CookieSecure, utils.CookieSameSite(s.Config.CookieSecure), csrfCookieName)
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func (s *Server) handleAuthMe(w http.ResponseWriter, r *http.Request) {
	claims, err := s.AuthFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	serverID := r.URL.Query().Get("server_id")
	authorized := true
	if serverID != "" {
		// Optional server authorization check.
		authorized, err = s.IsAuthorizedForServer(r.Context(), serverID, claims)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed auth check"})
			return
		}
	}

	csrfToken := ""
	if cookie, err := r.Cookie(csrfCookieName); err == nil {
		csrfToken = cookie.Value
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user": map[string]string{
			"id":        claims.UserID,
			"name":      claims.Name,
			"email":     claims.Email,
			"image":     claims.Image,
			"discordId": claims.DiscordID,
		},
		"authorized": authorized,
		"csrfToken":  csrfToken,
	})
}

func (s *Server) handleAuthServers(w http.ResponseWriter, r *http.Request) {
	claims, err := s.AuthFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	if claims.AccessToken == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing access token"})
		return
	}

	botGuilds, ok := botGuildMap()
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "bot not ready"})
		return
	}

	guilds, err := utils.FetchDiscordGuilds(r.Context(), claims.AccessToken)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to fetch guilds"})
		return
	}

	// Keep Discord's order, but only guilds the bot is actually in.
	present := make([]discordGuild, 0, len(guilds))
	for _, g := range guilds {
		if _, ok := botGuilds[g.ID]; ok {
			present = append(present, g)
		}
	}

	allowed, err := s.authorizedGuilds(r.Context(), present, claims)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "auth check failed"})
		return
	}

	servers := make([]serverSummary, 0, len(present))
	for _, g := range present {
		if !allowed[g.ID] {
			continue
		}
		servers = append(servers, serverSummary{
			ID:      g.ID,
			Name:    g.Name,
			IconURL: utils.GuildIconURL(g),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"servers": servers})
}

// AuthFromRequest extracts session claims from request cookie
func (s *Server) AuthFromRequest(r *http.Request) (*AuthClaims, error) {
	cookie, err := r.Cookie(authCookieName)
	if err != nil || cookie.Value == "" {
		return nil, errors.New("missing auth cookie")
	}

	session, err := s.DB.GetAuthSession(r.Context(), cookie.Value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("invalid session")
		}
		return nil, errors.New("failed to load session")
	}

	now := time.Now().Unix()
	if session.ExpiresAt <= now {
		_ = s.DB.DeleteAuthSession(r.Context(), session.SessionID)
		return nil, errors.New("session expired")
	}
	if session.LastSeen == 0 || now-session.LastSeen > int64(lastSeenGap.Seconds()) {
		_ = s.DB.UpdateAuthSessionLastSeen(r.Context(), session.SessionID, now)
	}

	accessToken := session.AccessToken
	if accessToken != "" {
		if strings.HasPrefix(accessToken, utils.TokenEncPrefix) {
			decrypted, err := utils.DecryptTokenV1(s.Config.AccessTokenKey, accessToken)
			if err != nil {
				return nil, errors.New("failed to decrypt session token")
			}
			accessToken = decrypted
		} else {
			encrypted, err := utils.EncryptTokenV1(s.Config.AccessTokenKey, accessToken)
			if err == nil {
				_ = s.DB.UpdateAuthSessionAccessToken(r.Context(), session.SessionID, encrypted)
			}
		}
	}

	return &AuthClaims{
		UserID:      session.UserID,
		Name:        session.Name,
		Email:       session.Email,
		Image:       session.Image,
		DiscordID:   session.DiscordID,
		AccessToken: accessToken,
	}, nil
}

// IsAuthorizedForServer checks if user is authorized to access a server
// Optimized: Uses bot cache + DB only, no Discord user API calls
func (s *Server) IsAuthorizedForServer(ctx context.Context, serverID string, claims *AuthClaims) (bool, error) {
	if !BotInGuild(serverID) {
		return false, nil
	}

	// 1. Check Discord Admin Cache first (0 DB queries for 99% of admin users!)
	if hasAdminPermsFromCache(serverID, claims.DiscordID) {
		return true, nil
	}

	id, err := strconv.ParseInt(serverID, 10, 64)
	if err != nil {
		return false, err
	}

	// 2. Check if configuration exists
	configExists := true
	_, err = s.DB.GetServerConfig(ctx, id)
	if err != nil {
		configExists = false
	}

	// 3. Check member whitelist
	isMember, _ := s.DB.IsMemberAuthorized(ctx, db.IsMemberAuthorizedParams{
		ServerConfigID: id,
		MemberID:       claims.DiscordID,
	})
	if isMember {
		return true, nil
	}

	if !configExists {
		return false, nil
	}

	// 4. Check role whitelist
	roleIds, _ := s.DB.GetAuthorizedRoles(ctx, id)
	if len(roleIds) > 0 && userHasAuthorizedRole(serverID, claims.DiscordID, roleIds) {
		return true, nil
	}

	return false, nil
}

func BotInGuild(guildID string) bool {
	if bot.Session == nil || bot.Session.State == nil {
		return false
	}
	bot.Session.State.RLock()
	defer bot.Session.State.RUnlock()
	_, err := bot.Session.State.Guild(guildID)
	return err == nil
}

func userHasAuthorizedRole(guildID, userID string, allowedRoles []string) bool {
	if bot.Session == nil || bot.Session.State == nil {
		return false
	}

	member, err := bot.Session.State.Member(guildID, userID)
	if err != nil {
		return false
	}

	for _, roleID := range member.Roles {
		if utils.ContainsString(allowedRoles, roleID) {
			return true
		}
	}
	return false
}

func hasAdminPermsFromCache(guildID, userID string) bool {
	if bot.Session == nil || bot.Session.State == nil {
		return false
	}

	guild, err := bot.Session.State.Guild(guildID)
	if err != nil {
		return false
	}

	if guild.OwnerID == userID {
		return true
	}

	member, err := bot.Session.State.Member(guildID, userID)
	if err != nil {
		return false
	}

	var perms int64
	for _, roleID := range member.Roles {
		role, err := bot.Session.State.Role(guildID, roleID)
		if err != nil {
			continue
		}
		perms |= role.Permissions
	}

	// Administrator = 0x8, Manage Guild = 0x20 only — MANAGE_CHANNELS intentionally excluded
	return perms&0x8 != 0 || perms&0x20 != 0
}

// authorizedGuilds resolves which of the given guilds the user may manage.
// Admin permissions come from the OAuth guild payload; whitelist lookups are
// batched into three queries per request instead of three per guild.
func (s *Server) authorizedGuilds(ctx context.Context, guilds []discordGuild, claims *AuthClaims) (map[string]bool, error) {
	allowed := make(map[string]bool, len(guilds))
	needsCheck := make([]discordGuild, 0, len(guilds))
	ids := make([]int64, 0, len(guilds))

	for _, g := range guilds {
		if utils.GuildIsAdmin(g) {
			allowed[g.ID] = true
			continue
		}
		id, err := strconv.ParseInt(g.ID, 10, 64)
		if err != nil {
			continue
		}
		needsCheck = append(needsCheck, g)
		ids = append(ids, id)
	}

	if len(needsCheck) == 0 {
		return allowed, nil
	}

	existing, err := s.DB.GetExistingServerConfigIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	configExists := make(map[int64]struct{}, len(existing))
	for _, id := range existing {
		configExists[id] = struct{}{}
	}

	memberServers, err := s.DB.GetServersByMember(ctx, claims.DiscordID)
	if err != nil {
		return nil, err
	}
	memberOf := make(map[int64]struct{}, len(memberServers))
	for _, id := range memberServers {
		memberOf[id] = struct{}{}
	}

	roleRows, err := s.DB.GetAuthorizedRolesForServers(ctx, ids)
	if err != nil {
		return nil, err
	}
	authorizedRoles := make(map[int64][]string, len(ids))
	for _, row := range roleRows {
		authorizedRoles[row.ServerConfigID] = append(authorizedRoles[row.ServerConfigID], row.RoleID)
	}

	for _, g := range needsCheck {
		id, _ := strconv.ParseInt(g.ID, 10, 64)
		_, hasConfig := configExists[id]
		_, isMember := memberOf[id]
		roleIds := authorizedRoles[id]
		cachedMatch := len(roleIds) > 0 && userHasAuthorizedRole(g.ID, claims.DiscordID, roleIds)

		allowed[g.ID] = guildAllowed(hasConfig, isMember, roleIds, cachedMatch, func() ([]string, error) {
			member, err := utils.FetchGuildMember(ctx, s.Config.BotToken, g.ID, claims.DiscordID)
			if err != nil {
				return nil, err
			}
			return member.Roles, nil
		})
	}

	return allowed, nil
}

// guildAllowed decides access for one non-admin guild from batched lookups.
// memberHasRole is only called on a Discord cache miss.
func guildAllowed(
	hasConfig bool,
	isMemberWhitelisted bool,
	roleIds []string,
	cachedRoleMatch bool,
	memberHasRole func() ([]string, error),
) bool {
	if !hasConfig {
		return false
	}
	if isMemberWhitelisted {
		return true
	}
	if len(roleIds) == 0 {
		return false
	}
	if cachedRoleMatch {
		return true
	}

	roles, err := memberHasRole()
	if err != nil {
		return false
	}
	for _, roleID := range roles {
		if utils.ContainsString(roleIds, roleID) {
			return true
		}
	}
	return false
}

// HasGuildPerms checks if user has admin/manage guild permissions on a guild
func (s *Server) HasGuildPerms(accessToken, guildID string) (bool, error) {
	guilds, err := utils.FetchDiscordGuilds(context.Background(), accessToken)
	if err != nil {
		return false, err
	}

	for _, g := range guilds {
		if g.ID != guildID {
			continue
		}
		if g.Owner {
			return true, nil
		}
		perms, err := strconv.ParseInt(g.Permissions, 10, 64)
		if err != nil {
			return false, err
		}
		if perms&utils.PermAdministrator != 0 || perms&utils.PermManageGuild != 0 {
			return true, nil
		}
		return false, nil
	}

	return false, nil
}

// MemberHasRole checks if user has any of the allowed roles in a guild
// Note: Using API call directly - member caching is complex in discordgo
func (s *Server) MemberHasRole(guildID, userID string, allowedRoles []string) (bool, error) {
	if s.Config.BotToken == "" {
		return false, errors.New("missing bot token")
	}

	member, err := utils.FetchGuildMember(context.Background(), s.Config.BotToken, guildID, userID)
	if err != nil {
		return false, err
	}

	for _, roleID := range member.Roles {
		if utils.ContainsString(allowedRoles, roleID) {
			return true, nil
		}
	}
	return false, nil
}

func botGuildMap() (map[string]struct{}, bool) {
	if bot.Session == nil || bot.Session.State == nil {
		return nil, false
	}

	state := bot.Session.State
	state.RLock()
	defer state.RUnlock()

	out := make(map[string]struct{}, len(state.Guilds))
	for _, g := range state.Guilds {
		out[g.ID] = struct{}{}
	}
	return out, true
}
