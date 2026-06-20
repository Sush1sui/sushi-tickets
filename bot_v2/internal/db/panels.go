package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const getPanelConfigsByServer = `
SELECT id, server_config_id, mention_roles_on_open, category_id, title, content, embed_color, channel_id, btn_color, btn_txt, btn_emoji, large_img_url, small_img_url
FROM panel_config
WHERE server_config_id = $1
ORDER BY id
`

func (q *Queries) GetPanelConfigsByServer(ctx context.Context, serverConfigID int64) ([]PanelConfig, error) {
	rows, err := q.db.Query(ctx, getPanelConfigsByServer, serverConfigID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]PanelConfig, 0)
	for rows.Next() {
		var i PanelConfig
		if err := rows.Scan(
			&i.ID,
			&i.ServerConfigID,
			&i.MentionRolesOnOpen,
			&i.CategoryID,
			&i.Title,
			&i.Content,
			&i.EmbedColor,
			&i.ChannelID,
			&i.BtnColor,
			&i.BtnTxt,
			&i.BtnEmoji,
			&i.LargeImgUrl,
			&i.SmallImgUrl,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getPanelConfigByID = `
SELECT id, server_config_id, mention_roles_on_open, category_id, title, content, embed_color, channel_id, btn_color, btn_txt, btn_emoji, large_img_url, small_img_url
FROM panel_config
WHERE server_config_id = $1 AND id = $2
`

func (q *Queries) GetPanelConfigByID(ctx context.Context, serverConfigID int64, panelID int32) (PanelConfig, error) {
	row := q.db.QueryRow(ctx, getPanelConfigByID, serverConfigID, panelID)
	var i PanelConfig
	err := row.Scan(
		&i.ID,
		&i.ServerConfigID,
		&i.MentionRolesOnOpen,
		&i.CategoryID,
		&i.Title,
		&i.Content,
		&i.EmbedColor,
		&i.ChannelID,
		&i.BtnColor,
		&i.BtnTxt,
		&i.BtnEmoji,
		&i.LargeImgUrl,
		&i.SmallImgUrl,
	)
	return i, err
}

type PanelButtonConfig struct {
	ID       int32
	Title    string
	BtnColor string
	BtnTxt   string
	BtnEmoji pgtype.Text
}

const getPanelButtonConfigsByIDs = `
SELECT id, title, btn_color, btn_txt, btn_emoji
FROM panel_config
WHERE server_config_id = $1
AND id = ANY($2)
`

func (q *Queries) GetPanelButtonConfigsByIDs(ctx context.Context, serverConfigID int64, panelIDs []int32) ([]PanelButtonConfig, error) {
	if len(panelIDs) == 0 {
		return []PanelButtonConfig{}, nil
	}

	rows, err := q.db.Query(ctx, getPanelButtonConfigsByIDs, serverConfigID, panelIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]PanelButtonConfig, 0)
	for rows.Next() {
		var i PanelButtonConfig
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.BtnColor,
			&i.BtnTxt,
			&i.BtnEmoji,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

type PanelQuestionsRow struct {
	Questions     []string
	RequiredFlags []bool
}

const getQuestionsByPanel = `
SELECT questions, required_flags
FROM questions_config
WHERE panel_config_id = $1
`

func (q *Queries) GetPanelQuestions(ctx context.Context, panelID int32) (PanelQuestionsRow, error) {
	var row PanelQuestionsRow
	err := q.db.QueryRow(ctx, getQuestionsByPanel, panelID).Scan(&row.Questions, &row.RequiredFlags)
	if err != nil {
		if err == pgx.ErrNoRows {
			return PanelQuestionsRow{}, nil
		}
		return PanelQuestionsRow{}, err
	}
	return row, nil
}

const getWelcomeByPanel = `
SELECT embed_color, title, description, title_url, large_img_url, small_img_url, footer, footer_icon_url
FROM welcome_msg_config
WHERE panel_config_id = $1
`

func (q *Queries) GetPanelWelcome(ctx context.Context, panelID int32) (WelcomeMsgConfig, bool, error) {
	var i WelcomeMsgConfig
	err := q.db.QueryRow(ctx, getWelcomeByPanel, panelID).Scan(
		&i.EmbedColor,
		&i.Title,
		&i.Description,
		&i.TitleUrl,
		&i.LargeImgUrl,
		&i.SmallImgUrl,
		&i.Footer,
		&i.FooterIconUrl,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return WelcomeMsgConfig{}, false, nil
		}
		return WelcomeMsgConfig{}, false, err
	}
	return i, true, nil
}

type CreatePanelConfigParams struct {
	ServerConfigID     int64
	MentionRolesOnOpen []string
	CategoryID         pgtype.Text
	Title              string
	Content            pgtype.Text
	EmbedColor         int32
	ChannelID          string
	BtnColor           string
	BtnTxt             string
	BtnEmoji           pgtype.Text
	LargeImgUrl        pgtype.Text
	SmallImgUrl        pgtype.Text
}

const createPanelConfig = `
INSERT INTO panel_config (
	server_config_id, mention_roles_on_open, category_id, title, content, embed_color, channel_id,
	btn_color, btn_txt, btn_emoji, large_img_url, small_img_url
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING id, server_config_id, mention_roles_on_open, category_id, title, content, embed_color, channel_id, btn_color, btn_txt, btn_emoji, large_img_url, small_img_url
`

func (q *Queries) CreatePanelConfig(ctx context.Context, arg CreatePanelConfigParams) (PanelConfig, error) {
	row := q.db.QueryRow(ctx, createPanelConfig,
		arg.ServerConfigID,
		arg.MentionRolesOnOpen,
		arg.CategoryID,
		arg.Title,
		arg.Content,
		arg.EmbedColor,
		arg.ChannelID,
		arg.BtnColor,
		arg.BtnTxt,
		arg.BtnEmoji,
		arg.LargeImgUrl,
		arg.SmallImgUrl,
	)
	var i PanelConfig
	err := row.Scan(
		&i.ID,
		&i.ServerConfigID,
		&i.MentionRolesOnOpen,
		&i.CategoryID,
		&i.Title,
		&i.Content,
		&i.EmbedColor,
		&i.ChannelID,
		&i.BtnColor,
		&i.BtnTxt,
		&i.BtnEmoji,
		&i.LargeImgUrl,
		&i.SmallImgUrl,
	)
	return i, err
}

type UpdatePanelConfigParams struct {
	ID                 int32
	ServerConfigID     int64
	MentionRolesOnOpen []string
	CategoryID         pgtype.Text
	Title              string
	Content            pgtype.Text
	EmbedColor         int32
	ChannelID          string
	BtnColor           string
	BtnTxt             string
	BtnEmoji           pgtype.Text
	LargeImgUrl        pgtype.Text
	SmallImgUrl        pgtype.Text
}

const updatePanelConfig = `
UPDATE panel_config
SET mention_roles_on_open = $3,
	category_id = $4,
	title = $5,
	content = $6,
	embed_color = $7,
	channel_id = $8,
	btn_color = $9,
	btn_txt = $10,
	btn_emoji = $11,
	large_img_url = $12,
	small_img_url = $13
WHERE id = $1 AND server_config_id = $2
RETURNING id, server_config_id, mention_roles_on_open, category_id, title, content, embed_color, channel_id, btn_color, btn_txt, btn_emoji, large_img_url, small_img_url
`

func (q *Queries) UpdatePanelConfig(ctx context.Context, arg UpdatePanelConfigParams) (PanelConfig, error) {
	row := q.db.QueryRow(ctx, updatePanelConfig,
		arg.ID,
		arg.ServerConfigID,
		arg.MentionRolesOnOpen,
		arg.CategoryID,
		arg.Title,
		arg.Content,
		arg.EmbedColor,
		arg.ChannelID,
		arg.BtnColor,
		arg.BtnTxt,
		arg.BtnEmoji,
		arg.LargeImgUrl,
		arg.SmallImgUrl,
	)
	var i PanelConfig
	err := row.Scan(
		&i.ID,
		&i.ServerConfigID,
		&i.MentionRolesOnOpen,
		&i.CategoryID,
		&i.Title,
		&i.Content,
		&i.EmbedColor,
		&i.ChannelID,
		&i.BtnColor,
		&i.BtnTxt,
		&i.BtnEmoji,
		&i.LargeImgUrl,
		&i.SmallImgUrl,
	)
	return i, err
}

const deletePanelConfig = `
DELETE FROM panel_config
WHERE id = $1 AND server_config_id = $2
`

func (q *Queries) DeletePanelConfig(ctx context.Context, id int32, serverConfigID int64) error {
	_, err := q.db.Exec(ctx, deletePanelConfig, id, serverConfigID)
	return err
}

const removePanelFromMultiPanels = `
UPDATE multi_panel_config
SET panel_config_ids = array_remove(panel_config_ids, $2)
WHERE server_config_id = $1
AND panel_config_ids @> ARRAY[$2]
`

func (q *Queries) RemovePanelFromMultiPanels(ctx context.Context, serverConfigID int64, panelID int32) error {
	_, err := q.db.Exec(ctx, removePanelFromMultiPanels, serverConfigID, panelID)
	return err
}

const deleteQuestionsByPanel = `
DELETE FROM questions_config
WHERE panel_config_id = $1
`

const createQuestionsConfig = `
INSERT INTO questions_config (panel_config_id, questions, required_flags)
VALUES ($1, $2, $3)
`

func (q *Queries) ReplaceQuestionsConfig(ctx context.Context, panelID int32, questions []string, requiredFlags []bool) error {
	if _, err := q.db.Exec(ctx, deleteQuestionsByPanel, panelID); err != nil {
		return err
	}
	if len(questions) == 0 {
		return nil
	}
	// Ensure requiredFlags length matches questions; pad with false if shorter.
	flags := make([]bool, len(questions))
	for i := range flags {
		if i < len(requiredFlags) {
			flags[i] = requiredFlags[i]
		}
	}
	_, err := q.db.Exec(ctx, createQuestionsConfig, panelID, questions, flags)
	return err
}

const deleteWelcomeByPanel = `
DELETE FROM welcome_msg_config
WHERE panel_config_id = $1
`

const createWelcomeConfig = `
INSERT INTO welcome_msg_config (
	panel_config_id, embed_color, title, description, title_url,
	large_img_url, small_img_url, footer, footer_icon_url
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`

type WelcomeMessageParams struct {
	EmbedColor    int32
	Title         string
	Description   string
	TitleURL      pgtype.Text
	LargeImgUrl   pgtype.Text
	SmallImgUrl   pgtype.Text
	Footer        pgtype.Text
	FooterIconUrl pgtype.Text
}

func (q *Queries) ReplaceWelcomeMsgConfig(ctx context.Context, panelID int32, params WelcomeMessageParams, enabled bool) error {
	if _, err := q.db.Exec(ctx, deleteWelcomeByPanel, panelID); err != nil {
		return err
	}
	if !enabled {
		return nil
	}
	_, err := q.db.Exec(
		ctx,
		createWelcomeConfig,
		panelID,
		params.EmbedColor,
		params.Title,
		params.Description,
		params.TitleURL,
		params.LargeImgUrl,
		params.SmallImgUrl,
		params.Footer,
		params.FooterIconUrl,
	)
	return err
}

const getMultiPanelConfigsByServer = `
SELECT id, server_config_id, title, content, embed_color, channel_id, large_img_url, small_img_url, use_dropdown, panel_config_ids, footer, foot_icon_url
FROM multi_panel_config
WHERE server_config_id = $1
ORDER BY id
`

func (q *Queries) GetMultiPanelConfigsByServer(ctx context.Context, serverConfigID int64) ([]MultiPanelConfig, error) {
	rows, err := q.db.Query(ctx, getMultiPanelConfigsByServer, serverConfigID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]MultiPanelConfig, 0)
	for rows.Next() {
		var i MultiPanelConfig
		if err := rows.Scan(
			&i.ID,
			&i.ServerConfigID,
			&i.Title,
			&i.Content,
			&i.EmbedColor,
			&i.ChannelID,
			&i.LargeImgUrl,
			&i.SmallImgUrl,
			&i.UseDropdown,
			&i.PanelConfigIds,
			&i.Footer,
			&i.FootIconUrl,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getMultiPanelConfigByID = `
SELECT id, server_config_id, title, content, embed_color, channel_id, large_img_url, small_img_url, use_dropdown, panel_config_ids, footer, foot_icon_url
FROM multi_panel_config
WHERE server_config_id = $1 AND id = $2
`

func (q *Queries) GetMultiPanelConfigByID(ctx context.Context, serverConfigID int64, multiPanelID int32) (MultiPanelConfig, error) {
	row := q.db.QueryRow(ctx, getMultiPanelConfigByID, serverConfigID, multiPanelID)
	var i MultiPanelConfig
	err := row.Scan(
		&i.ID,
		&i.ServerConfigID,
		&i.Title,
		&i.Content,
		&i.EmbedColor,
		&i.ChannelID,
		&i.LargeImgUrl,
		&i.SmallImgUrl,
		&i.UseDropdown,
		&i.PanelConfigIds,
		&i.Footer,
		&i.FootIconUrl,
	)
	return i, err
}

type CreateMultiPanelConfigParams struct {
	ServerConfigID int64
	Title          string
	Content        pgtype.Text
	EmbedColor     int32
	ChannelID      string
	LargeImgUrl    pgtype.Text
	SmallImgUrl    pgtype.Text
	UseDropdown    bool
	PanelConfigIds []int32
	Footer         pgtype.Text
	FootIconUrl    pgtype.Text
}

const createMultiPanelConfig = `
INSERT INTO multi_panel_config (
	server_config_id, title, content, embed_color, channel_id, large_img_url, small_img_url, use_dropdown, panel_config_ids, footer, foot_icon_url
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, server_config_id, title, content, embed_color, channel_id, large_img_url, small_img_url, use_dropdown, panel_config_ids, footer, foot_icon_url
`

func (q *Queries) CreateMultiPanelConfig(ctx context.Context, arg CreateMultiPanelConfigParams) (MultiPanelConfig, error) {
	row := q.db.QueryRow(ctx, createMultiPanelConfig,
		arg.ServerConfigID,
		arg.Title,
		arg.Content,
		arg.EmbedColor,
		arg.ChannelID,
		arg.LargeImgUrl,
		arg.SmallImgUrl,
		arg.UseDropdown,
		arg.PanelConfigIds,
		arg.Footer,
		arg.FootIconUrl,
	)
	var i MultiPanelConfig
	err := row.Scan(
		&i.ID,
		&i.ServerConfigID,
		&i.Title,
		&i.Content,
		&i.EmbedColor,
		&i.ChannelID,
		&i.LargeImgUrl,
		&i.SmallImgUrl,
		&i.UseDropdown,
		&i.PanelConfigIds,
		&i.Footer,
		&i.FootIconUrl,
	)
	return i, err
}

type UpdateMultiPanelConfigParams struct {
	ID             int32
	ServerConfigID int64
	Title          string
	Content        pgtype.Text
	EmbedColor     int32
	ChannelID      string
	LargeImgUrl    pgtype.Text
	SmallImgUrl    pgtype.Text
	UseDropdown    bool
	PanelConfigIds []int32
	Footer         pgtype.Text
	FootIconUrl    pgtype.Text
}

const updateMultiPanelConfig = `
UPDATE multi_panel_config
SET title = $3,
	content = $4,
	embed_color = $5,
	channel_id = $6,
	large_img_url = $7,
	small_img_url = $8,
	use_dropdown = $9,
	panel_config_ids = $10,
	footer = $11,
	foot_icon_url = $12
WHERE id = $1 AND server_config_id = $2
RETURNING id, server_config_id, title, content, embed_color, channel_id, large_img_url, small_img_url, use_dropdown, panel_config_ids, footer, foot_icon_url
`

func (q *Queries) UpdateMultiPanelConfig(ctx context.Context, arg UpdateMultiPanelConfigParams) (MultiPanelConfig, error) {
	row := q.db.QueryRow(ctx, updateMultiPanelConfig,
		arg.ID,
		arg.ServerConfigID,
		arg.Title,
		arg.Content,
		arg.EmbedColor,
		arg.ChannelID,
		arg.LargeImgUrl,
		arg.SmallImgUrl,
		arg.UseDropdown,
		arg.PanelConfigIds,
		arg.Footer,
		arg.FootIconUrl,
	)
	var i MultiPanelConfig
	err := row.Scan(
		&i.ID,
		&i.ServerConfigID,
		&i.Title,
		&i.Content,
		&i.EmbedColor,
		&i.ChannelID,
		&i.LargeImgUrl,
		&i.SmallImgUrl,
		&i.UseDropdown,
		&i.PanelConfigIds,
		&i.Footer,
		&i.FootIconUrl,
	)
	return i, err
}

const deleteMultiPanelConfig = `
DELETE FROM multi_panel_config
WHERE id = $1 AND server_config_id = $2
`

func (q *Queries) DeleteMultiPanelConfig(ctx context.Context, id int32, serverConfigID int64) error {
	_, err := q.db.Exec(ctx, deleteMultiPanelConfig, id, serverConfigID)
	return err
}
