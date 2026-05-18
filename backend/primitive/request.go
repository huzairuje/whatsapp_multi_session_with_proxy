package primitive

type SendTextSingleRequest struct {
	Recipient string `json:"recipient" validate:"required"`
	Message   string `json:"message" validate:"required"`
}

type SendTextBulkRequest struct {
	Recipients []string          `json:"recipients" validate:"required"`
	Message    string            `json:"message" validate:"required"`
	Variables  map[string]string `json:"variables,omitempty"`
}

type CheckUserSingleRequest struct {
	Recipient string `json:"recipient" validate:"required"`
}

type CheckUserBulkRequest struct {
	Recipients []string `json:"recipients" validate:"required"`
}

type SendSingleMediaRequest struct {
	Recipient string `json:"recipient" validate:"required"`
	File      string `json:"file" validate:"required"`
	Caption   string `json:"caption"`
}

type DeleteMessagesRequest struct {
	MessageIDs []string `json:"message_ids" validate:"required"`
	Recipient  string   `json:"recipient" validate:"required"`
}

type ConnectBulkDeviceRequest struct {
	Senders []string `json:"senders" validate:"required"`
}

type DisconnectBulkDeviceRequest struct {
	Senders []string `json:"senders" validate:"required"`
}
