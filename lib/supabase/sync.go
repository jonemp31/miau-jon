package supabase

import (
	"time"

	"go.uber.org/zap"
)

// InstanceData representa os dados COMPLETOS da instância para sincronização com Supabase
type InstanceData struct {
	// Identificação
	ID           string `json:"id"`
	InstanceName string `json:"instance_name"`
	RemoteJID    string `json:"remote_jid,omitempty"`
	PhoneNumber  string `json:"phone_number,omitempty"`

	// Status/Estado
	Status string `json:"status"`

	// Configurações de Comportamento
	RejectCall        bool   `json:"reject_call"`
	MsgCall           string `json:"msg_call,omitempty"`
	GroupsIgnore      bool   `json:"groups_ignore"`
	AlwaysOnline      bool   `json:"always_online"`
	ReadMessages      bool   `json:"read_messages"`
	ReadStatus        bool   `json:"read_status"`
	SyncFullHistory   bool   `json:"sync_full_history"`
	SyncRecentHistory bool   `json:"sync_recent_history"`

	// Webhook
	WebhookURL      string                 `json:"webhook_url,omitempty"`
	WebhookByEvents bool                   `json:"webhook_by_events"`
	WebhookBase64   bool                   `json:"webhook_base64"`
	WebhookHeaders  map[string]interface{} `json:"webhook_headers,omitempty"`
	WebhookEvents   []string               `json:"webhook_events,omitempty"`

	// Proxy
	ProxyHost     string `json:"proxy_host,omitempty"`
	ProxyPort     string `json:"proxy_port,omitempty"`
	ProxyProtocol string `json:"proxy_protocol,omitempty"`
	ProxyUsername string `json:"proxy_username,omitempty"`
	ProxyPassword string `json:"proxy_password,omitempty"`

	// Timestamps
	UpdatedAt      time.Time  `json:"updated_at"`
	ConnectedAt    *time.Time `json:"connected_at,omitempty"`
	LastSeenAt     *time.Time `json:"last_seen_at,omitempty"`
	DisconnectedAt *time.Time `json:"disconnected_at,omitempty"`

	// Analytics
	TotalMessagesSent     int        `json:"total_messages_sent,omitempty"`
	TotalMessagesReceived int        `json:"total_messages_received,omitempty"`
	ConnectionCount       int        `json:"connection_count,omitempty"`
	TotalErrors           int        `json:"total_errors,omitempty"`
	LastError             string     `json:"last_error,omitempty"`
	LastErrorAt           *time.Time `json:"last_error_at,omitempty"`

	// Metadados
	APIVersion string `json:"api_version,omitempty"`
	ServerIP   string `json:"server_ip,omitempty"`
	UserAgent  string `json:"user_agent,omitempty"`
	Platform   string `json:"platform,omitempty"`
}

// SyncInstance sincroniza os dados da instância com o Supabase de forma assíncrona
// Usa a tabela api-miau-v3
func SyncInstance(data InstanceData) {
	if !IsEnabled() {
		zap.L().Debug("Supabase sync skipped (not enabled)")
		return
	}

	// Garantir que tem timestamp
	if data.UpdatedAt.IsZero() {
		data.UpdatedAt = time.Now()
	}

	// Garantir versão da API
	if data.APIVersion == "" {
		data.APIVersion = "v3"
	}

	go func() {
		err := syncInstanceWithRetry(data, 3)
		if err != nil {
			zap.L().Error("Failed to sync instance to Supabase after retries",
				zap.String("instance_id", data.ID),
				zap.String("table", "api-miau-v3"),
				zap.Error(err),
			)
		} else {
			zap.L().Info("Instance synced to Supabase",
				zap.String("instance_id", data.ID),
				zap.String("status", data.Status),
				zap.String("table", "api-miau-v3"),
			)
		}
	}()
}

// syncInstanceWithRetry tenta sincronizar com retry automático
func syncInstanceWithRetry(data InstanceData, maxRetries int) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			// Backoff exponencial: 1s, 2s, 4s
			time.Sleep(time.Duration(1<<uint(i-1)) * time.Second)
			zap.L().Debug("Retrying Supabase sync",
				zap.String("instance_id", data.ID),
				zap.Int("attempt", i+1),
			)
		}

		err := Get().Upsert("api-miau-v3", data)
		if err == nil {
			return nil
		}

		lastErr = err
		zap.L().Warn("Supabase sync attempt failed",
			zap.String("instance_id", data.ID),
			zap.Int("attempt", i+1),
			zap.Error(err),
		)
	}

	return lastErr
}

// DeleteInstance remove a instância do Supabase
func DeleteInstance(instanceID string) {
	if !IsEnabled() {
		zap.L().Debug("Supabase delete skipped (not enabled)")
		return
	}

	go func() {
		err := Get().Delete("api-miau-v3", "id", instanceID)
		if err != nil {
			zap.L().Error("Failed to delete instance from Supabase",
				zap.String("instance_id", instanceID),
				zap.String("table", "api-miau-v3"),
				zap.Error(err),
			)
		} else {
			zap.L().Info("Instance deleted from Supabase",
				zap.String("instance_id", instanceID),
				zap.String("table", "api-miau-v3"),
			)
		}
	}()
}

// HealthCheck verifica se o Supabase está acessível
func HealthCheck() error {
	if !IsEnabled() {
		return nil
	}

	// Tenta fazer uma query simples na tabela api-miau-v3
	_, err := Get().Query("api-miau-v3", "limit=1")
	return err
}
