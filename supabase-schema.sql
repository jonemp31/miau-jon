-- ============================================
-- 📊 API-MIAU-V3 - TABELA COMPLETA SUPABASE
-- ============================================
-- Execute este SQL no Supabase SQL Editor
-- https://acersupa.painelopen.win/project/_/sql
-- ============================================

-- Criar tabela completa com todos os campos do Redis
CREATE TABLE "api-miau-v3" (
  -- ===== IDENTIFICAÇÃO =====
  id TEXT PRIMARY KEY,
  instance_name TEXT NOT NULL,
  remote_jid TEXT,
  phone_number TEXT,
  
  -- ===== STATUS/ESTADO =====
  status TEXT NOT NULL DEFAULT 'disconnected',
  
  -- ===== CONFIGURAÇÕES DE COMPORTAMENTO =====
  reject_call BOOLEAN DEFAULT false,
  msg_call TEXT,
  groups_ignore BOOLEAN DEFAULT false,
  always_online BOOLEAN DEFAULT false,
  read_messages BOOLEAN DEFAULT true,
  read_status BOOLEAN DEFAULT false,
  sync_full_history BOOLEAN DEFAULT false,
  sync_recent_history BOOLEAN DEFAULT false,
  
  -- ===== WEBHOOK =====
  webhook_url TEXT,
  webhook_by_events BOOLEAN DEFAULT false,
  webhook_base64 BOOLEAN DEFAULT false,
  webhook_headers JSONB,
  webhook_events JSONB,
  
  -- ===== PROXY (OPCIONAL) =====
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
  platform TEXT,
  
  -- ===== CONSTRAINTS =====
  CONSTRAINT valid_status CHECK (status IN ('open', 'connecting', 'close', 'disconnected', 'qr', 'pairing')),
  CONSTRAINT valid_proxy_protocol CHECK (proxy_protocol IS NULL OR proxy_protocol IN ('SOCKS5', 'HTTP', 'HTTPS'))
);

-- ===== ÍNDICES PARA PERFORMANCE =====
CREATE INDEX idx_api_miau_v3_status ON "api-miau-v3"(status);
CREATE INDEX idx_api_miau_v3_phone ON "api-miau-v3"(phone_number) WHERE phone_number IS NOT NULL;
CREATE INDEX idx_api_miau_v3_updated ON "api-miau-v3"(updated_at DESC);
CREATE INDEX idx_api_miau_v3_created ON "api-miau-v3"(created_at DESC);
CREATE INDEX idx_api_miau_v3_connected ON "api-miau-v3"(connected_at DESC) WHERE connected_at IS NOT NULL;

-- ===== TRIGGER PARA ATUALIZAR updated_at =====
CREATE OR REPLACE FUNCTION update_api_miau_v3_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER trigger_update_api_miau_v3_updated_at 
  BEFORE UPDATE ON "api-miau-v3"
  FOR EACH ROW 
  EXECUTE FUNCTION update_api_miau_v3_updated_at();

-- ===== RLS (Row Level Security) =====
ALTER TABLE "api-miau-v3" ENABLE ROW LEVEL SECURITY;

-- Policy para service_role ter acesso total
CREATE POLICY "service_role_all_access" ON "api-miau-v3"
  FOR ALL
  TO service_role
  USING (true)
  WITH CHECK (true);

-- ===== SUCESSO =====
SELECT 
  '✅ Tabela api-miau-v3 criada!' as status,
  COUNT(*) as total_colunas
FROM information_schema.columns 
WHERE table_name = 'api-miau-v3';
