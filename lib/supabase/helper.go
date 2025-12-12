package supabase

import (
	"strings"
	"time"

	"github.com/verbeux-ai/whatsmiau/models"
	"go.uber.org/zap"
)

// ConvertToInstanceData converte models.Instance para InstanceData do Supabase
func ConvertToInstanceData(instance *models.Instance, status string) InstanceData {
	// Extrair número de telefone limpo do RemoteJID
	phoneNumber := extractPhoneNumber(instance.RemoteJID)

	// Converter webhook events para array
	webhookEvents := []string{}
	if instance.Webhook.Events != nil {
		webhookEvents = instance.Webhook.Events
	}

	// Converter webhook headers para map[string]interface{}
	webhookHeaders := make(map[string]interface{})
	if instance.Webhook.Headers != nil {
		for k, v := range instance.Webhook.Headers {
			webhookHeaders[k] = v
		}
	}

	// Webhook by events
	webhookByEvents := false
	if instance.Webhook.ByEvents != nil {
		webhookByEvents = *instance.Webhook.ByEvents
	}

	// Webhook base64
	webhookBase64 := false
	if instance.Webhook.Base64 != nil {
		webhookBase64 = *instance.Webhook.Base64
	}

	data := InstanceData{
		// Identificação
		ID:           instance.ID,
		InstanceName: instance.ID,
		RemoteJID:    instance.RemoteJID,
		PhoneNumber:  phoneNumber,

		// Status
		Status: status,

		// Configurações
		RejectCall:        instance.RejectCall,
		MsgCall:           instance.MsgCall,
		GroupsIgnore:      instance.GroupsIgnore,
		AlwaysOnline:      instance.AlwaysOnline,
		ReadMessages:      instance.ReadMessages,
		ReadStatus:        instance.ReadStatus,
		SyncFullHistory:   instance.SyncFullHistory,
		SyncRecentHistory: instance.SyncRecentHistory,

		// Webhook
		WebhookURL:      instance.Webhook.Url,
		WebhookByEvents: webhookByEvents,
		WebhookBase64:   webhookBase64,
		WebhookHeaders:  webhookHeaders,
		WebhookEvents:   webhookEvents,

		// Proxy
		ProxyHost:     instance.ProxyHost,
		ProxyPort:     instance.ProxyPort,
		ProxyProtocol: instance.ProxyProtocol,
		ProxyUsername: instance.ProxyUsername,
		ProxyPassword: instance.ProxyPassword,

		// Timestamps
		UpdatedAt: time.Now(),

		// Metadados
		APIVersion: "v3",
	}

	// Definir connected_at se status for open
	if status == "open" && instance.RemoteJID != "" {
		now := time.Now()
		data.ConnectedAt = &now
		data.LastSeenAt = &now
	}

	return data
}

// extractPhoneNumber extrai o número limpo do JID
// Ex: "5521999999999@s.whatsapp.net" -> "5521999999999"
func extractPhoneNumber(remoteJID string) string {
	if remoteJID == "" {
		return ""
	}

	// Remove @s.whatsapp.net ou @g.us
	parts := strings.Split(remoteJID, "@")
	if len(parts) > 0 {
		return parts[0]
	}

	return remoteJID
}

// IncrementMessageSent incrementa contador de mensagens enviadas
func IncrementMessageSent(instanceID string) {
	if !IsEnabled() {
		return
	}

	go func() {
		// Chama RPC function do Supabase
		err := client.CallRPC("increment_messages_sent", map[string]interface{}{
			"instance_id": instanceID,
		})
		if err != nil {
			zap.L().Debug("failed to increment messages sent", zap.String("id", instanceID), zap.Error(err))
		}
	}()
}

// IncrementMessageReceived incrementa contador de mensagens recebidas
func IncrementMessageReceived(instanceID string) {
	if !IsEnabled() {
		return
	}

	go func() {
		// Chama RPC function do Supabase
		err := client.CallRPC("increment_messages_received", map[string]interface{}{
			"instance_id": instanceID,
		})
		if err != nil {
			zap.L().Debug("failed to increment messages received", zap.String("id", instanceID), zap.Error(err))
		}
	}()
}

// IncrementConnectionCount incrementa contador de conexões
func IncrementConnectionCount(instanceID string) {
	if !IsEnabled() {
		return
	}

	go func() {
		// Chama RPC function do Supabase
		err := client.CallRPC("increment_connection_count", map[string]interface{}{
			"instance_id": instanceID,
		})
		if err != nil {
			zap.L().Debug("failed to increment connection count", zap.String("id", instanceID), zap.Error(err))
		}
	}()
}

// RecordError registra erro na instância
func RecordError(instanceID string, errorMsg string) {
	if !IsEnabled() {
		return
	}

	go func() {
		// Chama RPC function do Supabase
		err := client.CallRPC("record_error", map[string]interface{}{
			"instance_id":   instanceID,
			"error_message": errorMsg,
		})
		if err != nil {
			zap.L().Debug("failed to record error", zap.String("id", instanceID), zap.Error(err))
		}
	}()
}
