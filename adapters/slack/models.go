package slack

type authTestResponse struct {
	OK     bool   `json:"ok"`
	URL    string `json:"url"`
	Team   string `json:"team"`
	TeamID string `json:"team_id"`
	User   string `json:"user"`
	UserID string `json:"user_id"`
	BotID  string `json:"bot_id"`
}

type slackErrorResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

type postMessageResponse struct {
	OK      bool        `json:"ok"`
	Channel string      `json:"channel"`
	TS      string      `json:"ts"`
	Message messageData `json:"message"`
}

type messageData struct {
	Type     string `json:"type"`
	Team     any    `json:"team"`
	Channel  any    `json:"channel"`
	TS       string `json:"ts"`
	ThreadTS string `json:"thread_ts,omitempty"`
	User     any    `json:"user"`
	Text     any    `json:"text"`
}

type conversationsListResponse struct {
	OK               bool             `json:"ok"`
	Channels         []channelData    `json:"channels"`
	ResponseMetadata responseMetadata `json:"response_metadata"`
}

type channelData struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsChannel bool   `json:"is_channel"`
	IsMember  bool   `json:"is_member"`
}

type responseMetadata struct {
	NextCursor string `json:"next_cursor"`
}
