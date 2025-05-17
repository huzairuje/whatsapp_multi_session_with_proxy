package primitive

type Message struct {
	MessageID string
	Jid       string
	Type      string
	Body      string
	Sent      bool
	FileName  string
}

type UpsertMessage struct {
	MessageID           string `json:"message_id"`
	Type                int    `json:"type"`
	IsSentUpsertMessage bool   `json:"isSentUpsertMessage"`
}
