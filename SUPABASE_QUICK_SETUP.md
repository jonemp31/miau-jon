# 🚀 GUIA RÁPIDO - Criar Tabela api-miau-v3 no Supabase

## ✅ Passos para Executar

### 1️⃣ Acesse o Supabase SQL Editor

Abra no navegador:
```
https://acersupa.painelopen.win/project/_/sql
```

### 2️⃣ Cole o SQL

Copie TODO o conteúdo do arquivo `supabase-schema.sql` e cole no editor SQL.

### 3️⃣ Execute

Clique no botão **RUN** (ou pressione Ctrl+Enter)

### 4️⃣ Aguarde a Confirmação

Você verá as mensagens:
```
✅ Tabela api-miau-v3 criada com sucesso!
📊 Inclui: 39 colunas
🔍 Índices: 5 criados para performance
🔒 RLS habilitado com policies
📈 Views criadas: api-miau-v3_active, api-miau-v3_stats
```

---

## 📊 O Que Foi Criado

### Tabela Principal: `api-miau-v3`

**39 colunas** incluindo:
- ✅ Identificação (id, instance_name, remote_jid, phone_number)
- ✅ Status (status, timestamps)
- ✅ Configurações completas (reject_call, groups_ignore, read_messages, etc)
- ✅ Webhook completo (url, events, headers, base64)
- ✅ Proxy completo (host, port, protocol, credentials)
- ✅ Analytics (messages sent/received, connections, errors)
- ✅ Metadados (api_version, server_ip, platform)

### Views Automáticas

1. **`api-miau-v3_active`** - Mostra só instâncias ativas (status=open)
2. **`api-miau-v3_stats`** - Estatísticas agregadas por status

### Índices de Performance

- Status
- Phone number
- Updated_at (DESC)
- Created_at (DESC)  
- Connected_at (DESC)

### Segurança (RLS)

- ✅ Service Role: Acesso total
- ✅ Anon Key: Somente leitura

---

## 🧪 Testar se Funcionou

No SQL Editor, execute:

```sql
-- Ver estrutura da tabela
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'api-miau-v3'
ORDER BY ordinal_position;

-- Ver se está vazia
SELECT COUNT(*) FROM "api-miau-v3";

-- Ver views
SELECT * FROM "api-miau-v3_stats";
```

---

## ⏭️ Próximos Passos

Após criar a tabela com sucesso:

1. ✅ Tabela criada
2. ⏳ Adicionar hooks nos controllers
3. ⏳ Testar sincronização
4. ⏳ Deploy da v3

---

**Status Atual**: ⏳ Aguardando você criar a tabela no Supabase
**Próximo**: Me confirme quando criar e vamos adicionar os hooks!
