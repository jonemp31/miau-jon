-- ============================================
-- 📊 LEADS-API-MIAU-V3 - TABELA DE LEADS
-- ============================================
-- Execute este SQL no Supabase SQL Editor
-- https://acersupa.painelopen.win/project/_/sql
-- ============================================

-- Criar tabela de leads
CREATE TABLE "leads-api-miau-v3" (
  -- ===== IDENTIFICAÇÃO =====
  id BIGSERIAL PRIMARY KEY,
  numero TEXT NOT NULL,                    -- remoteJid do contato
  nome TEXT,                               -- pushName do contato
  instancia TEXT NOT NULL,                 -- ID da instância que recebeu o contato
  
  -- ===== FLUXO E ETAPAS =====
  fluxo TEXT DEFAULT 'copy1',              -- Fluxo atual (default: copy1)
  etapa TEXT DEFAULT 'optin',              -- Etapa atual (default: optin)
  entregar TEXT DEFAULT 'copy1',           -- Controle de entrega
  
  -- ===== TIMESTAMPS =====
  optin TIMESTAMP WITH TIME ZONE DEFAULT NOW(),           -- Data primeira mensagem (cadastro)
  ultimo_envio TIMESTAMP WITH TIME ZONE,                  -- Última vez que API enviou mensagem
  ultima_interacao TIMESTAMP WITH TIME ZONE DEFAULT NOW(), -- Última mensagem do lead
  mensagem_status_hora TIMESTAMP WITH TIME ZONE,          -- Hora do último status
  
  -- ===== COMPRAS (N8N PREENCHE) =====
  comprou1 BOOLEAN DEFAULT false,
  comprou2 BOOLEAN DEFAULT false,
  comprou3 BOOLEAN DEFAULT false,
  comprou4 BOOLEAN DEFAULT false,
  comprou5 BOOLEAN DEFAULT false,
  
  -- ===== ÚLTIMA MENSAGEM DO LEAD =====
  ultima_msg_lead TEXT,                    -- Tipo: conversation, audioMessage, imageMessage, etc
  conteudo_msg TEXT,                       -- Texto da mensagem (message.conversation)
  
  -- ===== STATUS DA MENSAGEM =====
  mensagens_status TEXT,                   -- READ, DELIVERY_ACK, etc
  
  -- ===== CONSTRAINTS =====
  CONSTRAINT unique_numero_instancia UNIQUE (numero, instancia),
  CONSTRAINT valid_mensagens_status CHECK (
    mensagens_status IS NULL OR 
    mensagens_status IN ('READ', 'DELIVERY_ACK', 'SERVER_ACK', 'PLAYED', 'PENDING', 'ERROR')
  )
);

-- ===== ÍNDICES PARA PERFORMANCE =====
CREATE INDEX idx_leads_numero ON "leads-api-miau-v3"(numero);
CREATE INDEX idx_leads_instancia ON "leads-api-miau-v3"(instancia);
CREATE INDEX idx_leads_fluxo ON "leads-api-miau-v3"(fluxo);
CREATE INDEX idx_leads_etapa ON "leads-api-miau-v3"(etapa);
CREATE INDEX idx_leads_optin ON "leads-api-miau-v3"(optin DESC);
CREATE INDEX idx_leads_ultima_interacao ON "leads-api-miau-v3"(ultima_interacao DESC);
CREATE INDEX idx_leads_compras ON "leads-api-miau-v3"(comprou1, comprou2, comprou3, comprou4, comprou5) WHERE (comprou1 OR comprou2 OR comprou3 OR comprou4 OR comprou5);

-- ===== TRIGGER PARA ATUALIZAR ultima_interacao =====
CREATE OR REPLACE FUNCTION update_leads_ultima_interacao()
RETURNS TRIGGER AS $$
BEGIN
    -- Atualizar ultima_interacao sempre que houver UPDATE
    IF TG_OP = 'UPDATE' AND (OLD.conteudo_msg IS DISTINCT FROM NEW.conteudo_msg OR OLD.ultima_msg_lead IS DISTINCT FROM NEW.ultima_msg_lead) THEN
        NEW.ultima_interacao = NOW();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_leads_interacao
  BEFORE UPDATE ON "leads-api-miau-v3"
  FOR EACH ROW 
  EXECUTE FUNCTION update_leads_ultima_interacao();

-- ===== RLS (Row Level Security) =====
ALTER TABLE "leads-api-miau-v3" ENABLE ROW LEVEL SECURITY;

-- Policy para service_role ter acesso total
CREATE POLICY "service_role_all_access_leads" ON "leads-api-miau-v3"
  FOR ALL
  TO service_role
  USING (true)
  WITH CHECK (true);

-- ===== COMENTÁRIOS =====
COMMENT ON TABLE "leads-api-miau-v3" IS 'Armazena leads que interagem com as instâncias WhatsApp';
COMMENT ON COLUMN "leads-api-miau-v3".numero IS 'remoteJid do contato (ex: 5521999999999@s.whatsapp.net)';
COMMENT ON COLUMN "leads-api-miau-v3".nome IS 'pushName do contato';
COMMENT ON COLUMN "leads-api-miau-v3".optin IS 'Data da primeira mensagem (cadastro do lead)';
COMMENT ON COLUMN "leads-api-miau-v3".ultimo_envio IS 'Data/hora do último envio da API para o lead';
COMMENT ON COLUMN "leads-api-miau-v3".ultima_interacao IS 'Data/hora da última mensagem recebida do lead';
COMMENT ON COLUMN "leads-api-miau-v3".ultima_msg_lead IS 'Tipo da última mensagem: conversation, audioMessage, imageMessage, etc';
COMMENT ON COLUMN "leads-api-miau-v3".conteudo_msg IS 'Texto da mensagem (message.conversation)';
COMMENT ON COLUMN "leads-api-miau-v3".mensagens_status IS 'Status: READ, DELIVERY_ACK, SERVER_ACK, PLAYED, PENDING, ERROR';

-- ===== SUCESSO =====
SELECT 
  '✅ Tabela leads-api-miau-v3 criada!' as status,
  COUNT(*) as total_colunas
FROM information_schema.columns 
WHERE table_name = 'leads-api-miau-v3';

-- ===== QUERIES ÚTEIS =====
-- Ver leads recentes
-- SELECT id, numero, nome, instancia, fluxo, etapa, optin, ultima_interacao FROM "leads-api-miau-v3" ORDER BY ultima_interacao DESC LIMIT 10;

-- Leads por instância
-- SELECT instancia, COUNT(*) as total FROM "leads-api-miau-v3" GROUP BY instancia;

-- Leads que compraram
-- SELECT * FROM "leads-api-miau-v3" WHERE (comprou1 OR comprou2 OR comprou3 OR comprou4 OR comprou5);

-- Leads por fluxo e etapa
-- SELECT fluxo, etapa, COUNT(*) as total FROM "leads-api-miau-v3" GROUP BY fluxo, etapa;
