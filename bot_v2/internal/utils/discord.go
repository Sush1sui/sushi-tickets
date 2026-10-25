package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const DiscordAPIBase = "https://discord.com/api/v10"

// discordHTTPClient bounds how long a stalled Discord REST call can hold a request.
var discordHTTPClient = &http.Client{Timeout: 10 * time.Second}

type DiscordTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	Scope       string `json:"scope"`
}

type DiscordUser struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	GlobalName string `json:"global_name"`
	Email      string `json:"email"`
	Avatar     string `json:"avatar"`
}

type DiscordGuild struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Owner       bool   `json:"owner"`
	Permissions string `json:"permissions"`
}

type DiscordGuildMember struct {
	Roles []string `json:"roles"`
}

type ServerSummary struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	IconURL string `json:"iconUrl"`
}

const (
	PermAdministrator = 1 << 3
	PermManageGuild   = 1 << 5
)

func ExchangeDiscordToken(ctx context.Context, clientID, clientSecret, redirectURI, code string) (*DiscordTokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, DiscordAPIBase+"/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := discordHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var out DiscordTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func FetchDiscordUser(ctx context.Context, accessToken string) (*DiscordUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, DiscordAPIBase+"/users/@me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := discordHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fetch user failed: %s", string(body))
	}

	var out DiscordUser
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func FetchDiscordGuilds(ctx context.Context, accessToken string) ([]DiscordGuild, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, DiscordAPIBase+"/users/@me/guilds", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := discordHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fetch guilds failed: %s", string(body))
	}

	var out []DiscordGuild
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func FetchGuildMember(ctx context.Context, botToken, guildID, userID string) (*DiscordGuildMember, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, DiscordAPIBase+"/guilds/"+guildID+"/members/"+userID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bot "+botToken)

	resp, err := discordHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fetch member failed: %s", string(body))
	}

	var out DiscordGuildMember
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func GuildIconURL(guild DiscordGuild) string {
	if guild.Icon == "" {
		return ""
	}
	return fmt.Sprintf("https://cdn.discordapp.com/icons/%s/%s.png", guild.ID, guild.Icon)
}

func GuildIsAdmin(guild DiscordGuild) bool {
	if guild.Owner {
		return true
	}
	perms, err := ParsePermissions(guild.Permissions)
	if err != nil {
		return false
	}
	return perms&PermAdministrator != 0 || perms&PermManageGuild != 0
}

func ParsePermissions(permStr string) (int64, error) {
	return strconv.ParseInt(permStr, 10, 64)
}

func HasAdminPerms(perms int64) bool {
	return perms&PermAdministrator != 0 || perms&PermManageGuild != 0
}

type DiscordChannel struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     int    `json:"type"`
	ParentID string `json:"parent_id"`
}

type DiscordRole struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Color    int    `json:"color"`
	Position int    `json:"position"`
}

type DiscordEmoji struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Animated bool   `json:"animated"`
}

type DiscordMember struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	GlobalName string `json:"global_name"`
	Avatar     string `json:"avatar"`
	Bot        bool   `json:"bot"`
}

// ChannelTypeText is the type for text channels
const ChannelTypeText = 0
const ChannelTypeVoice = 2
const ChannelTypeCategory = 4

// GetGuildChannels fetches all channels for a guild via bot API
func GetGuildChannels(session *discordgo.Session, guildID string) ([]DiscordChannel, error) {
	channels, err := session.GuildChannels(guildID)
	if err != nil {
		return nil, err
	}

	result := make([]DiscordChannel, 0, len(channels))
	for _, ch := range channels {
		result = append(result, DiscordChannel{
			ID:       ch.ID,
			Name:     ch.Name,
			Type:     int(ch.Type),
			ParentID: ch.ParentID,
		})
	}
	return result, nil
}

// GetGuildChannelsCache gets channels from bot's state cache (like discord.js .channels.cache)
func GetGuildChannelsCache(session *discordgo.Session, guildID string) ([]DiscordChannel, bool) {
	if session == nil || session.State == nil {
		return nil, false
	}

	session.State.RLock()
	defer session.State.RUnlock()

	guild, err := session.State.Guild(guildID)
	if err != nil {
		return nil, false
	}

	result := make([]DiscordChannel, 0, len(guild.Channels))
	for _, ch := range guild.Channels {
		result = append(result, DiscordChannel{
			ID:       ch.ID,
			Name:     ch.Name,
			Type:     int(ch.Type),
			ParentID: ch.ParentID,
		})
	}
	return result, true
}

func GetGuildRolesCache(session *discordgo.Session, guildID string) ([]DiscordRole, bool) {
	if session == nil || session.State == nil {
		return nil, false
	}

	session.State.RLock()
	defer session.State.RUnlock()

	guild, err := session.State.Guild(guildID)
	if err != nil {
		return nil, false
	}

	result := make([]DiscordRole, 0, len(guild.Roles))
	for _, role := range guild.Roles {
		result = append(result, DiscordRole{
			ID:       role.ID,
			Name:     role.Name,
			Color:    int(role.Color),
			Position: role.Position,
		})
	}
	return result, true
}

func GetGuildEmojisCache(session *discordgo.Session, guildID string) ([]DiscordEmoji, bool) {
	if session == nil || session.State == nil {
		return nil, false
	}

	session.State.RLock()
	defer session.State.RUnlock()

	guild, err := session.State.Guild(guildID)
	if err != nil {
		return nil, false
	}

	result := make([]DiscordEmoji, 0, len(guild.Emojis))
	for _, emoji := range guild.Emojis {
		result = append(result, DiscordEmoji{
			ID:       emoji.ID,
			Name:     emoji.Name,
			Animated: emoji.Animated,
		})
	}
	return result, true
}

func GetGuildMembersCache(session *discordgo.Session, guildID string) ([]DiscordMember, bool) {
	if session == nil || session.State == nil {
		return nil, false
	}

	session.State.RLock()
	defer session.State.RUnlock()

	guild, err := session.State.Guild(guildID)
	if err != nil {
		return nil, false
	}

	result := make([]DiscordMember, 0, len(guild.Members))
	for _, m := range guild.Members {
		name := m.User.GlobalName
		if name == "" {
			name = m.User.Username
		}
		result = append(result, DiscordMember{
			ID:         m.User.ID,
			Username:   m.User.Username,
			GlobalName: name,
			Avatar:     m.User.Avatar,
			Bot:        m.User.Bot,
		})
	}
	return result, true
}
