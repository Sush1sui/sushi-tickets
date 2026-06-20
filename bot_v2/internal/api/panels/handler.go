package panels

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Sush1sui/FNS_BOT/internal/bot"
	"github.com/Sush1sui/FNS_BOT/internal/db"
	"github.com/Sush1sui/FNS_BOT/internal/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func validatePanelPayload(p PanelPayload) utils.ValidationErrors {
	errs := make(utils.ValidationErrors)
	utils.ValidateMaxLength(p.Title, "title", 256, errs)
	utils.ValidateMaxLength(p.Content, "content", 4000, errs)
	utils.ValidateSnowflake(p.ChannelID, "channelId", errs)
	utils.ValidateBtnColor(p.BtnColor, "btnColor", errs)
	utils.ValidateMaxLength(p.BtnTxt, "btnTxt", 80, errs)
	utils.ValidateMaxLength(p.BtnEmoji, "btnEmoji", 100, errs)
	utils.ValidateSnowflake(p.CategoryID, "categoryId", errs)
	utils.ValidateMaxLength(p.LargeImgUrl, "largeImgUrl", 2048, errs)
	utils.ValidateHTTPSURL(p.LargeImgUrl, "largeImgUrl", errs)
	utils.ValidateMaxLength(p.SmallImgUrl, "smallImgUrl", 2048, errs)
	utils.ValidateHTTPSURL(p.SmallImgUrl, "smallImgUrl", errs)
	if len(p.MentionRolesOnOpen) > 50 {
		errs["mentionRolesOnOpen"] = "max 50 roles allowed"
	} else {
		for i, roleID := range p.MentionRolesOnOpen {
			utils.ValidateSnowflake(roleID, "mentionRolesOnOpen["+strconv.Itoa(i)+"]", errs)
		}
	}
	if len(p.Questions) > 10 {
		errs["questions"] = "max 10 questions allowed"
	}
	for i, q := range p.Questions {
		utils.ValidateMaxLength(q.Label, "questions["+strconv.Itoa(i)+"].label", 256, errs)
	}
	utils.ValidateMaxLength(p.WelcomeMessage.Title, "welcomeMessage.title", 256, errs)
	utils.ValidateMaxLength(p.WelcomeMessage.Description, "welcomeMessage.description", 4000, errs)
	utils.ValidateMaxLength(p.WelcomeMessage.TitleURL, "welcomeMessage.titleUrl", 2048, errs)
	utils.ValidateHTTPSURL(p.WelcomeMessage.TitleURL, "welcomeMessage.titleUrl", errs)
	utils.ValidateMaxLength(p.WelcomeMessage.LargeImgUrl, "welcomeMessage.largeImgUrl", 2048, errs)
	utils.ValidateHTTPSURL(p.WelcomeMessage.LargeImgUrl, "welcomeMessage.largeImgUrl", errs)
	utils.ValidateMaxLength(p.WelcomeMessage.SmallImgUrl, "welcomeMessage.smallImgUrl", 2048, errs)
	utils.ValidateHTTPSURL(p.WelcomeMessage.SmallImgUrl, "welcomeMessage.smallImgUrl", errs)
	utils.ValidateMaxLength(p.WelcomeMessage.FooterText, "welcomeMessage.footerText", 2048, errs)
	utils.ValidateMaxLength(p.WelcomeMessage.FooterIconUrl, "welcomeMessage.footerIconUrl", 2048, errs)
	utils.ValidateHTTPSURL(p.WelcomeMessage.FooterIconUrl, "welcomeMessage.footerIconUrl", errs)
	return errs
}

func validateMultiPanelPayload(p MultiPanelPayload) utils.ValidationErrors {
	errs := make(utils.ValidationErrors)
	utils.ValidateMaxLength(p.Title, "title", 256, errs)
	utils.ValidateMaxLength(p.Content, "content", 4000, errs)
	utils.ValidateSnowflake(p.ChannelID, "channelId", errs)
	utils.ValidateMaxLength(p.LargeImgUrl, "largeImgUrl", 2048, errs)
	utils.ValidateHTTPSURL(p.LargeImgUrl, "largeImgUrl", errs)
	utils.ValidateMaxLength(p.SmallImgUrl, "smallImgUrl", 2048, errs)
	utils.ValidateHTTPSURL(p.SmallImgUrl, "smallImgUrl", errs)
	utils.ValidateMaxLength(p.Footer, "footer", 2048, errs)
	utils.ValidateMaxLength(p.FootIconUrl, "footIconUrl", 2048, errs)
	utils.ValidateHTTPSURL(p.FootIconUrl, "footIconUrl", errs)
	if len(p.PanelConfigIds) > 25 {
		errs["panelConfigIds"] = "max 25 panels allowed"
	}
	return errs
}



func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) HandleListPanels(w http.ResponseWriter, r *http.Request) {
	serverID, err := utils.ParseServerID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid server id"})
		return
	}

	items, err := h.DB.GetPanelConfigsByServer(context.Background(), serverID)
	if err != nil {
		log.Printf("load panels failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load panels"})
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) HandleGetPanel(w http.ResponseWriter, r *http.Request) {
	serverID, panelID, err := utils.ParseServerPanelIDs(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ids"})
		return
	}

	item, err := h.DB.GetPanelConfigByID(context.Background(), serverID, panelID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "panel not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load panel"})
		return
	}

	questions, err := h.DB.GetPanelQuestions(context.Background(), panelID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load questions"})
		return
	}

	// Map PanelQuestionsRow to []QuestionItem
	questionItems := make([]QuestionItem, len(questions.Questions))
	for i, label := range questions.Questions {
		req := false
		if i < len(questions.RequiredFlags) {
			req = questions.RequiredFlags[i]
		}
		questionItems[i] = QuestionItem{Label: label, IsRequired: req}
	}

	welcome, hasWelcome, err := h.DB.GetPanelWelcome(context.Background(), panelID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load welcome message"})
		return
	}

	var welcomePayload *WelcomeMessagePayload
	if hasWelcome {
		welcomePayload = &WelcomeMessagePayload{
			EmbedColor:    welcome.EmbedColor,
			Title:         welcome.Title,
			Description:   welcome.Description,
			TitleURL:      utils.TextOrEmpty(welcome.TitleUrl),
			LargeImgUrl:   utils.TextOrEmpty(welcome.LargeImgUrl),
			SmallImgUrl:   utils.TextOrEmpty(welcome.SmallImgUrl),
			FooterText:    utils.TextOrEmpty(welcome.Footer),
			FooterIconUrl: utils.TextOrEmpty(welcome.FooterIconUrl),
		}
	}

	writeJSON(w, http.StatusOK, PanelDetail{
		MentionRolesOnOpen: item.MentionRolesOnOpen,
		CategoryID:         utils.TextOrEmpty(item.CategoryID),
		Title:              item.Title,
		Content:            utils.TextOrEmpty(item.Content),
		EmbedColor:         item.EmbedColor,
		ChannelID:          item.ChannelID,
		BtnColor:           item.BtnColor,
		BtnTxt:             item.BtnTxt,
		BtnEmoji:           utils.TextOrEmpty(item.BtnEmoji),
		LargeImgUrl:        utils.TextOrEmpty(item.LargeImgUrl),
		SmallImgUrl:        utils.TextOrEmpty(item.SmallImgUrl),
		Questions:          questionItems,
		WelcomeMessage:     welcomePayload,
	})
}

func (h *Handler) HandleCreatePanel(w http.ResponseWriter, r *http.Request) {
	serverID, err := utils.ParseServerID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid server id"})
		return
	}

	if _, err := h.DB.GetServerConfig(context.Background(), serverID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if err := h.DB.EnsureServerConfig(context.Background(), serverID); err != nil {
				log.Printf("create default server config failed: %v", err)
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create server config"})
				return
			}
		} else {
			log.Printf("load server config failed: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load server config"})
			return
		}
	}

	var payload PanelPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	if missing := utils.MissingPanelFields(payload.Title, payload.Content, payload.ChannelID, payload.BtnColor, payload.BtnTxt); len(missing) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "missing required fields",
			"missing": strings.Join(missing, ", "),
		})
		return
	}

	if errs := validatePanelPayload(payload); len(errs) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "validation failed", "fields": errs})
		return
	}

	item, err := h.DB.CreatePanelConfig(context.Background(), db.CreatePanelConfigParams{
		ServerConfigID:     serverID,
		MentionRolesOnOpen: payload.MentionRolesOnOpen,
		CategoryID:         utils.ToText(payload.CategoryID),
		Title:              payload.Title,
		Content:            utils.ToText(payload.Content),
		EmbedColor:         payload.EmbedColor,
		ChannelID:          payload.ChannelID,
		BtnColor:           payload.BtnColor,
		BtnTxt:             payload.BtnTxt,
		BtnEmoji:           utils.ToText(payload.BtnEmoji),
		LargeImgUrl:        utils.ToText(payload.LargeImgUrl),
		SmallImgUrl:        utils.ToText(payload.SmallImgUrl),
	})
	if err != nil {
		log.Printf("create panel failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create panel"})
		return
	}
	qLabels := make([]string, 0, len(payload.Questions))
	qFlags := make([]bool, 0, len(payload.Questions))
	for _, q := range payload.Questions {
		if q.Label != "" {
			qLabels = append(qLabels, q.Label)
			qFlags = append(qFlags, q.IsRequired)
		}
	}
	if err := h.DB.ReplaceQuestionsConfig(context.Background(), item.ID, qLabels, qFlags); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save questions"})
		return
	}
	if err := h.DB.ReplaceWelcomeMsgConfig(
		context.Background(),
		item.ID,
		toWelcomeParams(payload.WelcomeMessage),
		utils.HasWelcomeMessage(
			payload.WelcomeMessage.Title,
			payload.WelcomeMessage.Description,
			payload.WelcomeMessage.TitleURL,
			payload.WelcomeMessage.LargeImgUrl,
			payload.WelcomeMessage.SmallImgUrl,
			payload.WelcomeMessage.FooterText,
			payload.WelcomeMessage.FooterIconUrl,
		),
	); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save welcome message"})
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *Handler) HandleUpdatePanel(w http.ResponseWriter, r *http.Request) {
	serverID, panelID, err := utils.ParseServerPanelIDs(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ids"})
		return
	}

	var payload PanelPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	if missing := utils.MissingPanelFields(payload.Title, payload.Content, payload.ChannelID, payload.BtnColor, payload.BtnTxt); len(missing) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "missing required fields",
			"missing": strings.Join(missing, ", "),
		})
		return
	}

	if errs := validatePanelPayload(payload); len(errs) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "validation failed", "fields": errs})
		return
	}

	item, err := h.DB.UpdatePanelConfig(context.Background(), db.UpdatePanelConfigParams{
		ID:                 panelID,
		ServerConfigID:     serverID,
		MentionRolesOnOpen: payload.MentionRolesOnOpen,
		CategoryID:         utils.ToText(payload.CategoryID),
		Title:              payload.Title,
		Content:            utils.ToText(payload.Content),
		EmbedColor:         payload.EmbedColor,
		ChannelID:          payload.ChannelID,
		BtnColor:           payload.BtnColor,
		BtnTxt:             payload.BtnTxt,
		BtnEmoji:           utils.ToText(payload.BtnEmoji),
		LargeImgUrl:        utils.ToText(payload.LargeImgUrl),
		SmallImgUrl:        utils.ToText(payload.SmallImgUrl),
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update panel"})
		return
	}
	uqLabels := make([]string, 0, len(payload.Questions))
	uqFlags := make([]bool, 0, len(payload.Questions))
	for _, q := range payload.Questions {
		if q.Label != "" {
			uqLabels = append(uqLabels, q.Label)
			uqFlags = append(uqFlags, q.IsRequired)
		}
	}
	if err := h.DB.ReplaceQuestionsConfig(context.Background(), panelID, uqLabels, uqFlags); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save questions"})
		return
	}
	if err := h.DB.ReplaceWelcomeMsgConfig(
		context.Background(),
		panelID,
		toWelcomeParams(payload.WelcomeMessage),
		utils.HasWelcomeMessage(
			payload.WelcomeMessage.Title,
			payload.WelcomeMessage.Description,
			payload.WelcomeMessage.TitleURL,
			payload.WelcomeMessage.LargeImgUrl,
			payload.WelcomeMessage.SmallImgUrl,
			payload.WelcomeMessage.FooterText,
			payload.WelcomeMessage.FooterIconUrl,
		),
	); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save welcome message"})
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) HandleDeletePanel(w http.ResponseWriter, r *http.Request) {
	serverID, panelID, err := utils.ParseServerPanelIDs(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ids"})
		return
	}

	if err := h.DB.RemovePanelFromMultiPanels(context.Background(), serverID, panelID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to prune multi panels"})
		return
	}
	if err := h.DB.DeletePanelConfig(context.Background(), panelID, serverID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete panel"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *Handler) HandleSendPanel(w http.ResponseWriter, r *http.Request) {
	serverID, panelID, err := utils.ParseServerPanelIDs(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ids"})
		return
	}

	sess := bot.Session
	if sess == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "bot not ready"})
		return
	}

	item, err := h.DB.GetPanelConfigByID(r.Context(), serverID, panelID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "panel not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load panel"})
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       item.Title,
		Description: utils.TextOrEmpty(item.Content),
		Color:       int(item.EmbedColor),
	}

	if item.LargeImgUrl.Valid {
		embed.Image = &discordgo.MessageEmbedImage{URL: item.LargeImgUrl.String}
	}
	if item.SmallImgUrl.Valid {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: item.SmallImgUrl.String}
	}

	button := discordgo.Button{
		Label:    item.BtnTxt,
		Style:    utils.ButtonStyleFromColor(item.BtnColor),
		CustomID: "open_ticket_" + strconv.Itoa(int(item.ID)),
	}
	if emoji := utils.ParseComponentEmoji(utils.TextOrEmpty(item.BtnEmoji)); emoji != nil {
		button.Emoji = emoji
	}

	msg, err := sess.ChannelMessageSendComplex(item.ChannelID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{embed},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{Components: []discordgo.MessageComponent{button}},
		},
	})
	if err != nil {
		log.Printf("send panel failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to send panel"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"messageId": msg.ID})
}

func (h *Handler) HandleListMultiPanels(w http.ResponseWriter, r *http.Request) {
	serverID, err := utils.ParseServerID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid server id"})
		return
	}

	items, err := h.DB.GetMultiPanelConfigsByServer(context.Background(), serverID)
	if err != nil {
		log.Printf("load multi panels failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load multi panels"})
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) HandleGetMultiPanel(w http.ResponseWriter, r *http.Request) {
	serverID, multiPanelID, err := utils.ParseServerMultiPanelIDs(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ids"})
		return
	}

	item, err := h.DB.GetMultiPanelConfigByID(context.Background(), serverID, multiPanelID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "multi panel not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load multi panel"})
		return
	}

	writeJSON(w, http.StatusOK, MultiPanelDetail{
		Title:          item.Title,
		Content:        utils.TextOrEmpty(item.Content),
		EmbedColor:     item.EmbedColor,
		ChannelID:      item.ChannelID,
		LargeImgUrl:    utils.TextOrEmpty(item.LargeImgUrl),
		SmallImgUrl:    utils.TextOrEmpty(item.SmallImgUrl),
		UseDropdown:    item.UseDropdown,
		PanelConfigIds: item.PanelConfigIds,
		Footer:         utils.TextOrEmpty(item.Footer),
		FootIconUrl:    utils.TextOrEmpty(item.FootIconUrl),
	})
}

func (h *Handler) HandleCreateMultiPanel(w http.ResponseWriter, r *http.Request) {
	serverID, err := utils.ParseServerID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid server id"})
		return
	}

	var payload MultiPanelPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	if missing := utils.MissingMultiPanelFields(payload.Title, payload.Content, payload.ChannelID); len(missing) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "missing required fields",
			"missing": strings.Join(missing, ", "),
		})
		return
	}

	if errs := validateMultiPanelPayload(payload); len(errs) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "validation failed", "fields": errs})
		return
	}

	item, err := h.DB.CreateMultiPanelConfig(context.Background(), db.CreateMultiPanelConfigParams{
		ServerConfigID: serverID,
		Title:          payload.Title,
		Content:        utils.ToText(payload.Content),
		EmbedColor:     payload.EmbedColor,
		ChannelID:      payload.ChannelID,
		LargeImgUrl:    utils.ToText(payload.LargeImgUrl),
		SmallImgUrl:    utils.ToText(payload.SmallImgUrl),
		UseDropdown:    payload.UseDropdown,
		PanelConfigIds: payload.PanelConfigIds,
		Footer:         utils.ToText(payload.Footer),
		FootIconUrl:    utils.ToText(payload.FootIconUrl),
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create multi panel"})
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *Handler) HandleUpdateMultiPanel(w http.ResponseWriter, r *http.Request) {
	serverID, multiPanelID, err := utils.ParseServerMultiPanelIDs(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ids"})
		return
	}

	var payload MultiPanelPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	if missing := utils.MissingMultiPanelFields(payload.Title, payload.Content, payload.ChannelID); len(missing) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "missing required fields",
			"missing": strings.Join(missing, ", "),
		})
		return
	}

	if errs := validateMultiPanelPayload(payload); len(errs) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "validation failed", "fields": errs})
		return
	}

	item, err := h.DB.UpdateMultiPanelConfig(context.Background(), db.UpdateMultiPanelConfigParams{
		ID:             multiPanelID,
		ServerConfigID: serverID,
		Title:          payload.Title,
		Content:        utils.ToText(payload.Content),
		EmbedColor:     payload.EmbedColor,
		ChannelID:      payload.ChannelID,
		LargeImgUrl:    utils.ToText(payload.LargeImgUrl),
		SmallImgUrl:    utils.ToText(payload.SmallImgUrl),
		UseDropdown:    payload.UseDropdown,
		PanelConfigIds: payload.PanelConfigIds,
		Footer:         utils.ToText(payload.Footer),
		FootIconUrl:    utils.ToText(payload.FootIconUrl),
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update multi panel"})
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) HandleDeleteMultiPanel(w http.ResponseWriter, r *http.Request) {
	serverID, multiPanelID, err := utils.ParseServerMultiPanelIDs(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ids"})
		return
	}

	if err := h.DB.DeleteMultiPanelConfig(context.Background(), multiPanelID, serverID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete multi panel"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *Handler) HandleSendMultiPanel(w http.ResponseWriter, r *http.Request) {
	serverID, multiPanelID, err := utils.ParseServerMultiPanelIDs(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ids"})
		return
	}

	sess := bot.Session
	if sess == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "bot not ready"})
		return
	}

	item, err := h.DB.GetMultiPanelConfigByID(r.Context(), serverID, multiPanelID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "multi panel not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load multi panel"})
		return
	}

	if len(item.PanelConfigIds) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "multi panel has no panels"})
		return
	}

	panels, err := h.DB.GetPanelButtonConfigsByIDs(r.Context(), serverID, item.PanelConfigIds)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load panel buttons"})
		return
	}
	if len(panels) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no panels available"})
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       item.Title,
		Description: utils.TextOrEmpty(item.Content),
		Color:       int(item.EmbedColor),
	}
	if item.LargeImgUrl.Valid {
		embed.Image = &discordgo.MessageEmbedImage{URL: item.LargeImgUrl.String}
	}
	if item.SmallImgUrl.Valid {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: item.SmallImgUrl.String}
	}
	if item.Footer.Valid {
		embed.Footer = &discordgo.MessageEmbedFooter{Text: item.Footer.String}
		if item.FootIconUrl.Valid {
			embed.Footer.IconURL = item.FootIconUrl.String
		}
	}

	components := utils.BuildMultiPanelComponents(item.UseDropdown, panels)
	msg, err := sess.ChannelMessageSendComplex(item.ChannelID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})
	if err != nil {
		log.Printf("send multi panel failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to send multi panel"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"messageId": msg.ID})
}

func toWelcomeParams(msg WelcomeMessagePayload) db.WelcomeMessageParams {
	return utils.ToWelcomeParams(
		msg.EmbedColor,
		msg.Title,
		msg.Description,
		msg.TitleURL,
		msg.LargeImgUrl,
		msg.SmallImgUrl,
		msg.FooterText,
		msg.FooterIconUrl,
	)
}

