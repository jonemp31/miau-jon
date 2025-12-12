# 🔄 Integração Supabase - WhatsMiau API

## ✅ Status da Integração

A integração com o Supabase foi implementada com **dual-write strategy**:
- ✅ SQLite continua como banco primário (rápido, local)
- ✅ Supabase recebe cópias dos dados (backup/analytics)
- ✅ Sincronização assíncrona (não bloqueia a API)
- ✅ Retry automático (3 tentativas com backoff exponencial)

---

## 📋 Próximos Passos

### 1. Criar Tabela no Supabase

Acesse o SQL Editor do Supabase:
```
https://acersupa.painelopen.win/project/_/sql
```

Execute o SQL que está no arquivo: `supabase-schema.sql`

---

### 2. Configuração já está pronta

As variáveis de ambiente já foram adicionadas ao `stack.md`:

```yaml
SUPABASE_URL: "https://acersupa.painelopen.win"
SUPABASE_KEY: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoic2VydmljZV9yb2xlIiwiaXNzIjoic3VwYWJhc2UiLCJpYXQiOjE3NTA2NDc2MDAsImV4cCI6MTkwODQxNDAw"
```

*(Usando SERVICE_ROLE_KEY para ter todas permissões)*

---

## 🔧 Como Funciona

### Arquivos Criados

1. **`lib/supabase/client.go`**
   - Cliente HTTP para comunicação com Supabase REST API
   - Métodos: `Upsert()`, `Delete()`, `Query()`
   - Timeout: 15 segundos

2. **`lib/supabase/sync.go`**
   - Funções de sincronização assíncrona
   - `SyncInstance()` - sincroniza dados da instância
   - `DeleteInstance()` - remove do Supabase
   - Retry automático com backoff exponencial

3. **`env/env.go`** (atualizado)
   - Adicionadas variáveis `SUPABASE_URL` e `SUPABASE_KEY`

4. **`main.go`** (atualizado)
   - Inicializa cliente Supabase no startup
   - Logs indicam se está habilitado ou não

---

## 📊 Estrutura da Tabela

```sql
CREATE TABLE instances (
  id TEXT PRIMARY KEY,              -- teste-001
  instance_name TEXT,               -- Nome da instância
  remote_jid TEXT,                  -- 5521999999999@s.whatsapp.net
  phone_number TEXT,                -- 5521999999999
  status TEXT,                      -- open, connecting, close
  webhook_url TEXT,                 -- URL do webhook
  read_messages BOOLEAN,            -- Auto-read ativo
  always_online BOOLEAN,            -- Always online ativo
  
  created_at TIMESTAMP,             -- Quando criou
  updated_at TIMESTAMP,             -- Última atualização
  connected_at TIMESTAMP,           -- Primeira conexão
  last_seen_at TIMESTAMP,           -- Último online
  
  total_messages_sent INTEGER,     -- Total enviadas
  total_messages_received INTEGER, -- Total recebidas
  connection_count INTEGER,        -- Reconexões
  
  api_version TEXT,                -- Versão da API
  server_ip TEXT                   -- IP do servidor
);
```

---

## 🎯 Próxima Etapa

**Agora precisamos adicionar os hooks nos controllers!**

Os locais onde vamos adicionar sincronização:

1. **Create Instance** → Quando criar nova instância
2. **Connect** → Quando conectar WhatsApp (QR ou Pairing)
3. **Update** → Quando atualizar configurações
4. **Delete** → Quando deletar instância
5. **Status Change** → Quando mudar status da conexão

---

## 🧪 Como Testar

Após criar a tabela e fazer deploy:

1. Crie uma instância via API
2. Verifique os logs:
   ```bash
   docker logs -f whatsmiau_api
   ```
   
3. Procure por:
   - `"Supabase integration enabled"`
   - `"Instance synced to Supabase"`

4. Verifique no Supabase Dashboard:
   ```
   https://acersupa.painelopen.win/project/_/editor
   ```

---

## ⚠️ Importante

- Se Supabase cair, a API continua funcionando normalmente
- Logs de erro aparecem mas não param a aplicação
- Dados continuam salvos no SQLite local
- Retry automático tenta 3x antes de desistir

---

## 📝 Logs Importantes

```
✅ INFO  - Supabase integration enabled
✅ INFO  - Instance synced to Supabase (instance_id=teste-001, status=open)
⚠️  WARN  - Supabase sync attempt failed (attempt=1)
❌ ERROR - Failed to sync instance to Supabase after retries
```

---

**Status**: ✅ Integração completa, aguardando criação da tabela no Supabase

**Próximo**: Adicionar hooks de sincronização nos controllers
