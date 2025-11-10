package dto

type ReadMessagesRequest struct {
	InstanceID   string                    `param:"instance" validate:"required"`
	ReadMessages []ReadMessagesRequestItem `json:"readMessages" validate:"required,min=1"`
}
type ReadMessagesRequestItem struct {
	RemoteJid string `json:"remoteJid" validate:"required"`
	//FromMe    bool   `json:"fromMe"` ignored
	Sender string `json:"sender"` // required if group
	ID     string `json:"id" validate:"required"`
}

type SendPresenceRequestPresence string

const (
	PresenceComposing SendPresenceRequestPresence = "composing"
	PresenceAvailable SendPresenceRequestPresence = "available"
)

type SendPresenceRequestType string

const (
	PresenceTypeText  SendPresenceRequestType = "text"
	PresenceTypeAudio SendPresenceRequestType = "audio"
)

type SendChatPresenceRequest struct {
	InstanceID string                      `param:"instance" validate:"required"`
	Number     string                      `json:"number"`
	Delay      int                         `json:"delay,omitempty" validate:"omitempty,min=0,max=300000"`
	Presence   SendPresenceRequestPresence `json:"presence"`
	Type       SendPresenceRequestType     `json:"type"`
}

type SendChatPresenceResponse struct {
	Presence SendPresenceRequestPresence `json:"presence"`
}

type NumberExistsRequest struct {
	Numbers []string `json:"numbers"     validate:"required,min=1,dive,required"`
}

// DeleteChatRequest - Requisição para deletar um chat
type DeleteChatRequest struct {
	InstanceID string `param:"instance" validate:"required"`
	Number     string `json:"number" validate:"required"`
}

type DeleteChatResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ArchiveChatRequest - Requisição para arquivar/desarquivar um chat
type ArchiveChatRequest struct {
	InstanceID string `param:"instance" validate:"required"`
	Number     string `json:"number" validate:"required"`
	Archive    bool   `json:"archive"` // true = arquivar, false = desarquivar
}

type ArchiveChatResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Archive bool   `json:"archive"`
}
