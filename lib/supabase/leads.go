package supabase

import (
	"time"

	"go.uber.org/zap"
)

// LeadData representa os dados do lead para sincronização com Supabase
type LeadData struct {
	ID                  int64      `json:"id,omitempty"`
	Numero              string     `json:"numero"`
	Nome                string     `json:"nome,omitempty"`
	Instancia           string     `json:"instancia"`
	Fluxo               string     `json:"fluxo"`
	Etapa               string     `json:"etapa"`
	Entregar            string     `json:"entregar"`
	Optin               time.Time  `json:"optin,omitempty"`
	UltimoEnvio         *time.Time `json:"ultimo_envio,omitempty"`
	UltimaInteracao     time.Time  `json:"ultima_interacao,omitempty"`
	MensagemStatusHora  *time.Time `json:"mensagem_status_hora,omitempty"`
	Comprou1            bool       `json:"comprou1"`
	Comprou2            bool       `json:"comprou2"`
	Comprou3            bool       `json:"comprou3"`
	Comprou4            bool       `json:"comprou4"`
	Comprou5            bool       `json:"comprou5"`
	UltimaMsgLead       string     `json:"ultima_msg_lead,omitempty"`
	ConteudoMsg         string     `json:"conteudo_msg,omitempty"`
	MensagensStatus     string     `json:"mensagens_status,omitempty"`
}

// SyncLead sincroniza ou atualiza um lead no Supabase
func SyncLead(data LeadData) {
	if !IsEnabled() {
		zap.L().Debug("Supabase sync skipped (not enabled)")
		return
	}

	// Garantir valores padrão
	if data.Fluxo == "" {
		data.Fluxo = "copy1"
	}
	if data.Etapa == "" {
		data.Etapa = "optin"
	}
	if data.Entregar == "" {
		data.Entregar = "copy1"
	}
	if data.Optin.IsZero() {
		data.Optin = time.Now()
	}
	if data.UltimaInteracao.IsZero() {
		data.UltimaInteracao = time.Now()
	}

	go func() {
		err := syncLeadWithRetry(data, 3)
		if err != nil {
			zap.L().Error("Failed to sync lead to Supabase after retries",
				zap.String("numero", data.Numero),
				zap.String("instancia", data.Instancia),
				zap.String("table", "leads-api-miau-v3"),
				zap.Error(err),
			)
		} else {
			zap.L().Info("Lead synced to Supabase",
				zap.String("numero", data.Numero),
				zap.String("instancia", data.Instancia),
				zap.String("table", "leads-api-miau-v3"),
			)
			
			// Incrementar contador de leads na instância
			IncrementLeadsCount(data.Instancia)
		}
	}()
}

// syncLeadWithRetry tenta sincronizar o lead com retry automático
func syncLeadWithRetry(data LeadData, maxRetries int) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			time.Sleep(backoff)
			zap.L().Debug("Retrying lead sync",
				zap.Int("attempt", attempt),
				zap.Duration("backoff", backoff),
			)
		}

		err := client.Upsert("leads-api-miau-v3", data)
		if err == nil {
			return nil
		}

		lastErr = err
		zap.L().Warn("Lead sync attempt failed",
			zap.Int("attempt", attempt+1),
			zap.Error(err),
		)
	}

	return lastErr
}

// IncrementLeadsCount incrementa contador de leads na instância
func IncrementLeadsCount(instanceID string) {
	if !IsEnabled() {
		return
	}

	go func() {
		// Chama RPC function do Supabase
		err := client.CallRPC("increment_leads_count", map[string]interface{}{
			"instance_id": instanceID,
		})
		if err != nil {
			zap.L().Debug("failed to increment leads count", zap.String("id", instanceID), zap.Error(err))
		}
	}()
}

// UpdateLeadLastSent atualiza o último envio para o lead
func UpdateLeadLastSent(numero, instancia string) {
	if !IsEnabled() {
		return
	}

	go func() {
		now := time.Now()
		data := map[string]interface{}{
			"numero":       numero,
			"instancia":    instancia,
			"ultimo_envio": now,
		}

		err := client.Upsert("leads-api-miau-v3", data)
		if err != nil {
			zap.L().Debug("failed to update lead last sent",
				zap.String("numero", numero),
				zap.String("instancia", instancia),
				zap.Error(err))
		}
	}()
}
