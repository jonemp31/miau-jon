package dto

import "github.com/verbeux-ai/whatsmiau/models"

type CreateInstanceRequest struct {
	ID               string `json:"id,omitempty" validate:"required_without=InstanceName"`
	InstanceName     string `json:"instanceName,omitempty" validate:"required_without=InstanceID"`
	*models.Instance        // optional arguments
}

type CreateInstanceResponse struct {
	*models.Instance
}

type UpdateInstanceRequest struct {
	ID      string `json:"id,omitempty" param:"id" validate:"required"`
	Webhook struct {
		Base64 bool   `json:"base64,omitempty"`
		URL    string `json:"url,omitempty"`
	} `json:"webhook,omitempty"`
}

type UpdateInstanceResponse struct {
	*models.Instance
}

type ListInstancesRequest struct {
	InstanceName string `query:"instanceName"`
	ID           string `query:"id"`
}

type ListInstancesResponse struct {
	*models.Instance

	OwnerJID     string `json:"ownerJid,omitempty"`
	RemoteJid    string `json:"remoteJid,omitempty"` // Alias para compatibilidade
	InstanceName string `json:"instanceName,omitempty"`
	Status       string `json:"status,omitempty"`
}

type ConnectInstanceRequest struct {
	ID              string `param:"id" validate:"required"`
	FingerprintType string `json:"fingerprintType,omitempty"` // "chrome", "firefox", "safari", "edge" (default: chrome)
}

type ConnectInstanceResponse struct {
	Message   string `json:"message,omitempty"`
	Connected bool   `json:"connected,omitempty"`
	Base64    string `json:"base64,omitempty"`
	*models.Instance
}

type StatusInstanceRequest struct {
	ID string `param:"id" validate:"required"`
}

type StatusInstanceResponse struct {
	ID        string                                        `json:"id,omitempty"`
	Status    string                                        `json:"state,omitempty"`
	RemoteJid string                                        `json:"remoteJid,omitempty"`
	Instance  *StatusInstanceResponseEvolutionCompatibility `json:"instance,omitempty"`
}

type StatusInstanceResponseEvolutionCompatibility struct {
	InstanceName string `json:"instanceName,omitempty"`
	State        string `json:"state,omitempty"`
	RemoteJid    string `json:"remoteJid,omitempty"`
}

type DeleteInstanceRequest struct {
	ID string `param:"id" validate:"required"`
}

type DeleteInstanceResponse struct {
	Message string `json:"message,omitempty"`
}

type LogoutInstanceRequest struct {
	ID string `param:"id" validate:"required"`
}

type LogoutInstanceResponse struct {
	Message string `json:"message,omitempty"`
}

// PairInstanceRequest - Requisição para solicitar código de pareamento
type PairInstanceRequest struct {
	ID              string `param:"id" validate:"required"`
	PhoneNumber     string `json:"phoneNumber" validate:"required,min=10,max=15"` // Ex: 5511999999999
	FingerprintType string `json:"fingerprintType,omitempty"`                     // "chrome", "firefox", "safari", "edge" (default: chrome)
}

// PairInstanceResponse - Resposta com código de pareamento
type PairInstanceResponse struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// PairingStatusRequest - Requisição para verificar status do pareamento
type PairingStatusRequest struct {
	ID string `param:"id" validate:"required"`
}

// PairingStatusResponse - Resposta com status do pareamento
type PairingStatusResponse struct {
	Status string `json:"status"` // "waiting", "connected", "expired", "not_found"
	Code   string `json:"code,omitempty"`
}
