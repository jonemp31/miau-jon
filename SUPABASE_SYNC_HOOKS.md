# 🔄 Sincronização Automática Supabase

## 📊 Tabela: `api-miau-v3`

Todos os eventos de **criação, atualização, conexão e desconexão** de instâncias são automaticamente sincronizados com o Supabase em tempo real.

---

## ✅ Hooks Implementados

### 1️⃣ **Criação de Instância**
**Arquivo:** `server/controllers/instance.go` → `Create()`
```go
// Sincronizar com Supabase
supabaseData := supabase.ConvertToInstanceData(request.Instance, "disconnected")
supabase.SyncInstance(supabaseData)
```

**Evento:** POST `/v1/instance`  
**Status inicial:** `disconnected`  
**Campos sincronizados:** ID, webhook, configurações, timestamps

---

### 2️⃣ **Atualização de Configurações**
**Arquivo:** `server/controllers/instance.go` → `Update()`
```go
// Sincronizar com Supabase (buscar status atual)
status, _ := s.whatsmiau.Status(request.ID)
supabaseData := supabase.ConvertToInstanceData(instance, string(status))
supabase.SyncInstance(supabaseData)
```

**Evento:** PUT `/v1/instance/:id`  
**Mantém status atual:** `open`, `disconnected`, etc.  
**Campos atualizados:** Webhook URL, headers, eventos, base64, etc.

---

### 3️⃣ **Conexão via QR Code**
**Arquivo:** `lib/whatsmiau/whatsmeow.go` → `observeConnection()`
```go
// Sincronizar com Supabase (status open + connected_at)
if instanceFound := s.getInstance(id); instanceFound != nil {
    supabaseData := supabase.ConvertToInstanceData(instanceFound, "open")
    supabase.SyncInstance(supabaseData)
    zap.L().Info("instance connected via QR and synced to supabase", zap.String("id", id))
}
```

**Evento:** QR Code lido no WhatsApp  
**Status:** `disconnected` → `open`  
**Timestamps atualizados:** `connected_at`, `last_seen_at`  
**Campo preenchido:** `remote_jid` (número do WhatsApp)

---

### 4️⃣ **Conexão via Pairing Code**
**Arquivo:** `lib/whatsmiau/whatsmeow.go` → `observePairing()`
```go
// Sincronizar com Supabase (status open + connected_at)
supabaseData := supabase.ConvertToInstanceData(instance, "open")
supabase.SyncInstance(supabaseData)
zap.L().Info("instance connected via pairing and synced to supabase", zap.String("id", id))
```

**Evento:** Código de pareamento aceito no app móvel  
**Status:** `disconnected` → `open`  
**Timestamps atualizados:** `connected_at`, `last_seen_at`  
**Campo preenchido:** `remote_jid`, `phone_number`

---

### 5️⃣ **Desconexão / Logout**
**Arquivo:** `lib/whatsmiau/event_emitter.go` → `handleLoggedOut()`
```go
// Sincronizar status desconectado no Supabase
instance := s.getInstanceCached(id)
if instance != nil {
    supabaseData := supabase.ConvertToInstanceData(instance, "disconnected")
    supabase.SyncInstance(supabaseData)
    zap.L().Info("instance disconnected and synced to supabase", zap.String("id", id))
}
```

**Evento:** WhatsApp desconecta (logout, ban, sessão expirada)  
**Status:** `open` → `disconnected`  
**Timestamp atualizado:** `disconnected_at`  
**Campo limpo:** Mantém `remote_jid` para histórico

---

### 6️⃣ **Exclusão de Instância**
**Arquivo:** `server/controllers/instance.go` → `Delete()`
```go
// Remover do Supabase
supabase.DeleteInstance(request.ID)
zap.L().Info("instance deleted and removed from supabase", zap.String("id", request.ID))
```

**Evento:** DELETE `/v1/instance/:id`  
**Ação:** Remove linha da tabela `api-miau-v3`

---

## 🔧 Arquitetura de Sincronização

### **Dual-Write Strategy**
```
API Request
    ↓
SQLite (Primary DB) ✅ Fast, local
    ↓ (async)
Supabase (Backup/Analytics) ☁️ Cloud, não bloqueia
```

### **Função Helper**
**Arquivo:** `lib/supabase/helper.go`
```go
func ConvertToInstanceData(instance *models.Instance, status string) InstanceData
```

Converte o modelo Redis (`models.Instance`) para o formato Supabase (`InstanceData`) com 39 campos.

### **Retry Automático**
**Arquivo:** `lib/supabase/sync.go` → `syncInstanceWithRetry()`
```
Tentativa 1: Imediata
Tentativa 2: Após 1 segundo
Tentativa 3: Após 2 segundos
Tentativa 4: Após 4 segundos (backoff exponencial)
```

Se falhar 3 vezes: Apenas log de erro, **não bloqueia a API**.

---

## 📝 Logs de Sincronização

Todos os eventos geram logs estruturados:

```log
INFO  instance created and synced to supabase  {"id": "teste-001"}
INFO  instance updated and synced to supabase  {"id": "teste-001"}
INFO  instance connected via QR and synced to supabase  {"id": "teste-001"}
INFO  instance connected via pairing and synced to supabase  {"id": "teste-001"}
INFO  instance disconnected and synced to supabase  {"id": "teste-001"}
INFO  instance deleted and removed from supabase  {"id": "teste-001"}
```

**Erros de sincronização:**
```log
ERROR  failed to sync instance to Supabase after 3 retries  {"id": "teste-001", "error": "..."}
```

---

## 🎯 Campos Sincronizados (39 colunas)

### **Identificação**
- `id`, `instance_name`, `remote_jid`, `phone_number`

### **Status**
- `status` (open, connecting, close, disconnected, qr, pairing)

### **Configurações de Comportamento**
- `reject_call`, `msg_call`, `groups_ignore`
- `always_online`, `read_messages`, `read_status`
- `sync_full_history`, `sync_recent_history`

### **Webhook**
- `webhook_url`, `webhook_by_events`, `webhook_base64`
- `webhook_headers` (JSONB), `webhook_events` (JSONB)

### **Proxy**
- `proxy_host`, `proxy_port`, `proxy_protocol`
- `proxy_username`, `proxy_password`

### **Timestamps**
- `created_at` (auto: primeira inserção)
- `updated_at` (auto: trigger atualiza a cada mudança)
- `connected_at`, `last_seen_at`, `disconnected_at`

### **Analytics** ✨ (atualizado automaticamente)
- `total_messages_sent` ➕ Incrementado a cada mensagem enviada
- `total_messages_received` ➕ Incrementado a cada mensagem recebida
- `connection_count` ➕ Incrementado a cada nova conexão (QR ou Pairing)
- `total_errors` (preparado para futuro)
- `last_error`, `last_error_at` (preparado para futuro)

### **Metadados**
- `api_version` (fixo: "v3")
- `server_ip`, `user_agent`, `platform` (preparado para futuro)

---

## 📊 Como Funcionam os Analytics

### **Mensagens Enviadas/Recebidas**
**Quando atualiza:**  
- A cada mensagem processada pelo event handler
- Diferencia automaticamente: `IsFromMe = true` → enviada | `false` → recebida

**Implementação:**
```go
// event_emitter.go → handleMessageEvent()
if e.Info.IsFromMe {
    supabase.IncrementMessageSent(instance.ID)
} else {
    supabase.IncrementMessageReceived(instance.ID)
}
```

**SQL Function:**
```sql
CREATE FUNCTION increment_messages_sent(instance_id TEXT)
UPDATE "api-miau-v3" SET total_messages_sent = total_messages_sent + 1, updated_at = NOW()
```

### **Contador de Conexões**
**Quando atualiza:**  
- Cada vez que a instância estabelece conexão com WhatsApp
- Tanto via QR Code quanto via Pairing Code

**Implementação:**
```go
// whatsmeow.go → observeConnection() e observePairing()
supabase.IncrementConnectionCount(id)
```

**SQL Function:**
```sql
CREATE FUNCTION increment_connection_count(instance_id TEXT)
UPDATE "api-miau-v3" SET connection_count = connection_count + 1, updated_at = NOW()
```

### **Vantagens dos Analytics**
✅ **Atômico:** Incrementos SQL diretos (sem race conditions)  
✅ **Assíncrono:** Não bloqueia processamento de mensagens  
✅ **Resiliente:** Se falhar, apenas log (não quebra API)  
✅ **Performático:** Chamadas RPC otimizadas

---

## 🚀 Testar Sincronização

### 1. Criar instância:
```bash
POST /v1/instance
Body: {"instanceName": "teste-sync", "readMessages": true}
```
✅ Deve aparecer na tabela Supabase com `status = "disconnected"`

### 2. Conectar via QR Code:
```bash
POST /v1/instance/connect
Body: {"id": "teste-sync"}
```
Ler QR no WhatsApp → ✅ Status muda para `"open"`, `connected_at` preenchido

### 3. Atualizar webhook:
```bash
PUT /v1/instance/teste-sync
Body: {"webhook": {"url": "https://novo-webhook.com"}}
```
✅ Campo `webhook_url` atualizado no Supabase

### 4. Desconectar:
```bash
POST /v1/instance/logout
Body: {"id": "teste-sync"}
```
✅ Status muda para `"disconnected"`, `disconnected_at` preenchido

### 5. Deletar:
```bash
DELETE /v1/instance/teste-sync
Body: {"id": "teste-sync"}
```
✅ Linha removida da tabela Supabase

---

## 🔍 Consultar no Supabase

### **SQL Editor** (https://acersupa.painelopen.win/project/_/sql)

```sql
-- Ver todas as instâncias
SELECT id, status, phone_number, connected_at, updated_at 
FROM "api-miau-v3" 
ORDER BY updated_at DESC;

-- Ver apenas conectadas
SELECT * FROM "api-miau-v3" WHERE status = 'open';

-- Analytics: Total de instâncias por status
SELECT status, COUNT(*) as total 
FROM "api-miau-v3" 
GROUP BY status;

-- Instâncias criadas hoje
SELECT * FROM "api-miau-v3" 
WHERE created_at::date = CURRENT_DATE;

-- Últimas desconexões
SELECT id, phone_number, disconnected_at 
FROM "api-miau-v3" 
WHERE disconnected_at IS NOT NULL 
ORDER BY disconnected_at DESC 
LIMIT 10;
```

---

## ✅ Status da Implementação

- ✅ Tabela `api-miau-v3` criada (38 colunas)
- ✅ Cliente HTTP Supabase REST API (`lib/supabase/client.go`)
- ✅ Lógica de sincronização com retry (`lib/supabase/sync.go`)
- ✅ Função helper de conversão (`lib/supabase/helper.go`)
- ✅ Hook Create (controller)
- ✅ Hook Update (controller)
- ✅ Hook Delete (controller)
- ✅ Hook Conexão QR Code (whatsmeow)
- ✅ Hook Conexão Pairing (whatsmeow)
- ✅ Hook Desconexão (event_emitter)
- ✅ Inicialização automática no `main.go`
- ✅ Variáveis de ambiente (`SUPABASE_URL`, `SUPABASE_KEY`)

---

## 🎉 Pronto!

Todas as operações de instância agora sincronizam automaticamente com o Supabase.

**Benefícios:**
- 📊 **Backup automático** de todas as instâncias
- 📈 **Analytics** e relatórios via SQL
- 🌍 **Multi-servidor** (futuro)
- 🔍 **Auditoria** completa (histórico de conexões)
- 🚀 **Não bloqueia** a API (async + retry)
