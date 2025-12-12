# 📊 SQL Completo - API Miau v3.0 + Supabase

## 🎯 Instruções de Uso

### 1️⃣ Acessar o Supabase SQL Editor
- Acesse: https://acersupa.painelopen.win/project/_/sql
- Ou vá em: **SQL Editor** no menu lateral do Supabase

### 2️⃣ Executar o SQL abaixo
- Copie todo o conteúdo da seção "SQL COMPLETO"
- Cole no SQL Editor
- Clique em **RUN** ou pressione `Ctrl+Enter`

### 3️⃣ Configurar Variáveis de Ambiente

#### 🐳 Docker Compose (docker-compose.yml)
```yaml
services:
  whatsmiau:
    image: jondevsouza31/miau-jon:latest
    environment:
      - API_KEY=sua_api_key_aqui
      - SERVER_PORT=8097
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - SUPABASE_URL=https://acersupa.painelopen.win
      - SUPABASE_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImFjZXJzdXBhIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTYyMzM2MzYwMCwiZXhwIjoxOTM4OTM5NjAwfQ.exemplo
```

#### 📁 Arquivo .env
```env
# API Configuration
API_KEY=sua_api_key_aqui
SERVER_PORT=8097

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379

# Supabase Configuration (OPCIONAL)
SUPABASE_URL=https://acersupa.painelopen.win
SUPABASE_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImFjZXJzdXBhIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTYyMzM2MzYwMCwiZXhwIjoxOTM4OTM5NjAwfQ.exemplo
```

#### 🔑 Como obter SUPABASE_KEY
1. Acesse seu projeto no Supabase
2. Vá em **Settings** → **API**
3. Copie a **service_role key** (secret)
4. ⚠️ **IMPORTANTE**: Use a `service_role` key, NÃO a `anon` key

#### 🔍 Como obter SUPABASE_URL
1. Acesse seu projeto no Supabase
2. Vá em **Settings** → **API**
3. Copie a **Project URL**
4. Exemplo: `https://seu-projeto.supabase.co`

---

## 📦 SQL COMPLETO

Execute o código abaixo no Supabase SQL Editor:

```sql
-- ============================================
-- 📊 TABELA API-MIAU-V3
-- ============================================
-- Tabela principal para armazenar todas as instâncias WhatsApp
-- com sincronização automática de analytics e eventos
-- ============================================

CREATE TABLE IF NOT EXISTS "api-miau-v3" (
  -- ===== IDENTIFICAÇÃO =====
  id TEXT PRIMARY KEY,
  instance_name TEXT NOT NULL,
  remote_jid TEXT,
  phone_number TEXT,
  
  -- ===== STATUS =====
  status TEXT NOT NULL DEFAULT 'pending',
  
  -- ===== CONFIGURAÇÕES DE COMPORTAMENTO =====
  reject_call BOOLEAN DEFAULT false,
  msg_call TEXT,
  groups_ignore BOOLEAN DEFAULT false,
  always_online BOOLEAN DEFAULT false,
  read_messages BOOLEAN DEFAULT false,
  read_status BOOLEAN DEFAULT false,
  sync_full_history BOOLEAN DEFAULT false,
  sync_recent_history BOOLEAN DEFAULT false,
  
  -- ===== WEBHOOK =====
  webhook_url TEXT,
  webhook_by_events BOOLEAN DEFAULT false,
  webhook_base64 BOOLEAN DEFAULT false,
  webhook_headers JSONB,
  webhook_events TEXT[],
  
  -- ===== PROXY =====
  proxy_host TEXT,
  proxy_port TEXT,
  proxy_protocol TEXT,
  proxy_username TEXT,
  proxy_password TEXT,
  
  -- ===== TIMESTAMPS =====
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  connected_at TIMESTAMP WITH TIME ZONE,
  last_seen_at TIMESTAMP WITH TIME ZONE,
  disconnected_at TIMESTAMP WITH TIME ZONE,
  
  -- ===== ANALYTICS =====
  total_messages_sent INTEGER DEFAULT 0,
  total_messages_received INTEGER DEFAULT 0,
  connection_count INTEGER DEFAULT 0,
  total_errors INTEGER DEFAULT 0,
  last_error TEXT,
  last_error_at TIMESTAMP WITH TIME ZONE,
  
  -- ===== METADADOS =====
  api_version TEXT DEFAULT 'v3',
  server_ip TEXT,
  user_agent TEXT,
  platform TEXT
);

-- ===== ÍNDICES PARA PERFORMANCE =====
CREATE INDEX IF NOT EXISTS idx_api_miau_v3_instance_name ON "api-miau-v3"(instance_name);
CREATE INDEX IF NOT EXISTS idx_api_miau_v3_status ON "api-miau-v3"(status);
CREATE INDEX IF NOT EXISTS idx_api_miau_v3_phone_number ON "api-miau-v3"(phone_number);
CREATE INDEX IF NOT EXISTS idx_api_miau_v3_created_at ON "api-miau-v3"(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_api_miau_v3_updated_at ON "api-miau-v3"(updated_at DESC);

-- ===== TRIGGER PARA ATUALIZAR updated_at =====
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_api_miau_v3_updated_at ON "api-miau-v3";
CREATE TRIGGER update_api_miau_v3_updated_at
  BEFORE UPDATE ON "api-miau-v3"
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();

-- ===== ROW LEVEL SECURITY (RLS) =====
ALTER TABLE "api-miau-v3" ENABLE ROW LEVEL SECURITY;

-- Policy para permitir todas operações com service_role
DROP POLICY IF EXISTS "Enable all for service_role" ON "api-miau-v3";
CREATE POLICY "Enable all for service_role"
  ON "api-miau-v3"
  FOR ALL
  USING (auth.role() = 'service_role');

-- ============================================
-- 📊 FUNÇÕES RPC PARA ANALYTICS
-- ============================================
-- Funções para incrementar contadores de forma atômica
-- e segura, evitando race conditions
-- ============================================

-- Função: Incrementar mensagens enviadas
CREATE OR REPLACE FUNCTION increment_messages_sent(instance_id TEXT)
RETURNS VOID AS $$
BEGIN
  UPDATE "api-miau-v3"
  SET 
    total_messages_sent = total_messages_sent + 1,
    updated_at = NOW()
  WHERE id = instance_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função: Incrementar mensagens recebidas
CREATE OR REPLACE FUNCTION increment_messages_received(instance_id TEXT)
RETURNS VOID AS $$
BEGIN
  UPDATE "api-miau-v3"
  SET 
    total_messages_received = total_messages_received + 1,
    updated_at = NOW()
  WHERE id = instance_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função: Incrementar contador de conexões
CREATE OR REPLACE FUNCTION increment_connection_count(instance_id TEXT)
RETURNS VOID AS $$
BEGIN
  UPDATE "api-miau-v3"
  SET 
    connection_count = connection_count + 1,
    updated_at = NOW()
  WHERE id = instance_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função: Registrar erro
CREATE OR REPLACE FUNCTION record_error(instance_id TEXT, error_message TEXT)
RETURNS VOID AS $$
BEGIN
  UPDATE "api-miau-v3"
  SET 
    total_errors = total_errors + 1,
    last_error = error_message,
    last_error_at = NOW(),
    updated_at = NOW()
  WHERE id = instance_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- ============================================
-- ✅ VERIFICAÇÃO
-- ============================================

-- Verificar se as tabelas foram criadas
SELECT 'Tabelas criadas com sucesso!' AS status;

-- Listar tabelas
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = 'public' 
AND table_name = 'api-miau-v3';

-- Listar funções
SELECT routine_name 
FROM information_schema.routines 
WHERE routine_schema = 'public' 
AND routine_name LIKE 'increment%' OR routine_name = 'record_error';

-- ============================================
-- 📝 CONSULTAS ÚTEIS
-- ============================================

-- Ver todas as instâncias
-- SELECT * FROM "api-miau-v3" ORDER BY created_at DESC;

-- Ver instâncias conectadas
-- SELECT id, instance_name, status, connected_at, total_messages_sent, total_messages_received 
-- FROM "api-miau-v3" 
-- WHERE status = 'open' 
-- ORDER BY connected_at DESC;

-- Ver analytics de uma instância
-- SELECT 
--   id,
--   instance_name,
--   status,
--   total_messages_sent,
--   total_messages_received,
--   connection_count,
--   created_at,
--   connected_at
-- FROM "api-miau-v3" 
-- WHERE id = 'sua-instancia-id';
```

---

## 🔒 Segurança

### ⚠️ IMPORTANTE
- **NUNCA** compartilhe sua `service_role` key publicamente
- Use variáveis de ambiente para armazenar credenciais
- A `service_role` key tem acesso TOTAL ao banco de dados
- Em produção, adicione as chaves no gerenciador de secrets da sua stack

### 🔐 Boas Práticas
1. Use `.env` para desenvolvimento local
2. Use secrets do Docker/Kubernetes em produção
3. Rotacione as keys periodicamente
4. Monitore logs de acesso no Supabase

---

## 📈 Recursos

### ✅ Funcionalidades Ativas
- ✅ Sincronização automática de instâncias
- ✅ Analytics em tempo real (mensagens, conexões)
- ✅ Timestamps automáticos
- ✅ Backup automático na nuvem
- ✅ Retry automático com exponential backoff
- ✅ Dual-write strategy (SQLite + Supabase)
- ✅ RLS (Row Level Security) ativado
- ✅ Índices otimizados para performance

### 🚀 Vantagens do Supabase
- **Opcional**: API funciona normalmente sem Supabase
- **Real-time**: Atualizações em tempo real
- **Escalável**: Suporta milhares de instâncias
- **Confiável**: Backup automático na nuvem
- **Integrável**: Fácil integração com N8N, Make, Zapier

---

## 📞 Suporte

Se tiver dúvidas:
1. Verifique os logs da aplicação
2. Teste as queries SQL diretamente no Supabase
3. Confirme que as variáveis de ambiente estão corretas
4. Verifique se a `service_role` key está sendo usada

---

**Versão da API**: v3.0-supabase  
**Última atualização**: Dezembro 2024  
**Docker Image**: `jondevsouza31/miau-jon:latest`
