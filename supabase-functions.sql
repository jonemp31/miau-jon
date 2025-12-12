-- ============================================
-- 📊 FUNÇÕES SQL PARA INCREMENT NO SUPABASE
-- ============================================
-- Execute no Supabase SQL Editor
-- https://acersupa.painelopen.win/project/_/sql
-- ============================================

-- Função para incrementar mensagens enviadas
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

-- Função para incrementar mensagens recebidas
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

-- Função para incrementar contador de conexões
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

-- Função para registrar erro
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

-- ===== TESTAR =====
-- SELECT increment_messages_sent('teste-001');
-- SELECT * FROM "api-miau-v3" WHERE id = 'teste-001';
