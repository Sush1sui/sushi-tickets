package panels

import "github.com/Sush1sui/FNS_BOT/internal/db"

// QuestionItem represents a single panel question with style, placeholder, and required flag.
type QuestionItem struct {
	Label       string `json:"label"`
	IsRequired  bool   `json:"isRequired"`
	Style       string `json:"style,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
}

type Handler struct {
	DB *db.Queries
}

type PanelPayload struct {
	MentionRolesOnOpen  []string              `json:"mentionRolesOnOpen"`
	CategoryID          string                `json:"categoryId"`
	Title               string                `json:"title"`
	Content             string                `json:"content"`
	EmbedColor          int32                 `json:"embedColor"`
	ChannelID           string                `json:"channelId"`
	BtnColor            string                `json:"btnColor"`
	BtnTxt              string                `json:"btnTxt"`
	BtnEmoji            string                `json:"btnEmoji"`
	LargeImgUrl         string                `json:"largeImgUrl"`
	SmallImgUrl         string                `json:"smallImgUrl"`
	Questions           []QuestionItem        `json:"questions"`
	QuestionsModalTitle string                `json:"questionsModalTitle,omitempty"`
	WelcomeMessage      WelcomeMessagePayload `json:"welcomeMessage"`
}

type WelcomeMessagePayload struct {
	EmbedColor    int32  `json:"embedColor"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	TitleURL      string `json:"titleUrl"`
	LargeImgUrl   string `json:"largeImgUrl"`
	SmallImgUrl   string `json:"smallImgUrl"`
	FooterText    string `json:"footerText"`
	FooterIconUrl string `json:"footerIconUrl"`
}

type MultiPanelPayload struct {
	Title          string  `json:"title"`
	Content        string  `json:"content"`
	EmbedColor     int32   `json:"embedColor"`
	ChannelID      string  `json:"channelId"`
	LargeImgUrl    string  `json:"largeImgUrl"`
	SmallImgUrl    string  `json:"smallImgUrl"`
	UseDropdown    bool    `json:"useDropdown"`
	PanelConfigIds []int32 `json:"panelConfigIds"`
	Footer         string  `json:"footer"`
	FootIconUrl    string  `json:"footIconUrl"`
}

type PanelDetail struct {
	MentionRolesOnOpen  []string               `json:"mentionRolesOnOpen"`
	CategoryID          string                 `json:"categoryId"`
	Title               string                 `json:"title"`
	Content             string                 `json:"content"`
	EmbedColor          int32                  `json:"embedColor"`
	ChannelID           string                 `json:"channelId"`
	BtnColor            string                 `json:"btnColor"`
	BtnTxt              string                 `json:"btnTxt"`
	BtnEmoji            string                 `json:"btnEmoji"`
	LargeImgUrl         string                 `json:"largeImgUrl"`
	SmallImgUrl         string                 `json:"smallImgUrl"`
	Questions           []QuestionItem         `json:"questions"`
	QuestionsModalTitle string                 `json:"questionsModalTitle"`
	WelcomeMessage      *WelcomeMessagePayload `json:"welcomeMessage"`
}

type MultiPanelDetail struct {
	Title          string  `json:"title"`
	Content        string  `json:"content"`
	EmbedColor     int32   `json:"embedColor"`
	ChannelID      string  `json:"channelId"`
	LargeImgUrl    string  `json:"largeImgUrl"`
	SmallImgUrl    string  `json:"smallImgUrl"`
	UseDropdown    bool    `json:"useDropdown"`
	PanelConfigIds []int32 `json:"panelConfigIds"`
	Footer         string  `json:"footer"`
	FootIconUrl    string  `json:"footIconUrl"`
}
