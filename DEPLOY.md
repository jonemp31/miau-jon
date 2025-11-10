# 🐱 WhatsMiau - Docker Stack para Portainer

Stack completa do WhatsMiau com PostgreSQL e Redis otimizados para alta escalabilidade (100-300 conexões simultâneas).

## 📦 Componentes da Stack

- **WhatsMiau API** - API principal (porta 8097)
- **PostgreSQL 16** - Banco de dados com 200 conexões simultâneas
- **Redis 7** - Cache e storage com persistence

## 🚀 Deploy Rápido no Portainer

### 1️⃣ Criar Volumes

Antes de deploy da stack, crie os volumes externos:

```bash
docker volume create postgres_data_whatsmiau
docker volume create redis_data_whatsmiau
```

### 2️⃣ Deploy da Stack

1. Acesse seu Portainer
2. Vá em **Stacks** > **Add Stack**
3. Cole o conteúdo do `docker-compose.yml`
4. **IMPORTANTE:** Edite as variáveis de ambiente:
   - `DEFAULT_WEBHOOK_URL` - URL do seu webhook
   - `API_KEY` - Sua chave de API
   - Senhas do PostgreSQL e Redis (se desejar)

5. Clique em **Deploy the stack**

### 3️⃣ Verificar Health

Aguarde ~1 minuto e verifique:

```bash
# Ver status dos serviços
docker service ls | grep whatsmiau

# Ver logs
docker service logs whatsmiau_api --tail 50
```

## ⚙️ Configurações Principais

### 🔗 Webhook Obrigatório

Todas as instâncias criadas usarão este webhook por padrão:

```yaml
- DEFAULT_WEBHOOK_URL=http://192.168.100.149:5680/webhook/whatsmiau
```

### 📡 Filtros de Eventos

Controle quais eventos são enviados para o webhook:

```yaml
# Todos os eventos (padrão)
- DEFAULT_WEBHOOK_EVENTS=All

# Ou eventos específicos (separados por vírgula)
- DEFAULT_WEBHOOK_EVENTS=MESSAGES_UPSERT,MESSAGES_UPDATE,CONNECTION_UPDATE

# Filtros adicionais
- DEFAULT_SKIP_GROUPS=false          # Pular eventos de grupos
- DEFAULT_SKIP_BROADCASTS=false      # Pular broadcasts/status
- DEFAULT_SKIP_OWN_MESSAGES=false    # Pular mensagens próprias
```

**Eventos disponíveis:**
- `MESSAGES_UPSERT` - Mensagens recebidas
- `MESSAGES_UPDATE` - Status de entrega/leitura
- `MESSAGES_DELETE` - Mensagens deletadas
- `CONTACTS_UPSERT` - Contatos atualizados
- `CONNECTION_UPDATE` - Status de conexão
- `GROUPS_UPSERT` - Grupos atualizados
- `GROUP_PARTICIPANTS_UPDATE` - Participantes de grupo
- `CALL` - Chamadas (offer, accept, terminate)

### 🤖 Auto-Features

Recursos automáticos aplicados em novas instâncias:

```yaml
- DEFAULT_AUTO_RECEIPT=true       # 2 checks cinza automáticos
- DEFAULT_AUTO_READ=true          # Marca como lido após 8s (2 checks azuis)
- DEFAULT_READ_MESSAGES=true      # Habilita leitura
- DEFAULT_ALWAYS_ONLINE=false     # AlwaysOnline (15min interval)
- DEFAULT_REJECT_CALLS=false      # Rejeita chamadas
```

### 📊 Escalabilidade

Configurações para alta performance:

```yaml
- EMITTER_BUFFER_SIZE=10000       # Buffer de 10.000 eventos
- HANDLER_SEMAPHORE_SIZE=500      # 500 handlers simultâneos
```

**Recursos alocados:**
- API: 4 CPUs, 4GB RAM
- PostgreSQL: 2 CPUs, 2GB RAM, 200 conexões
- Redis: 1 CPU, 1.5GB RAM

## 📝 Endpoints da API

Base URL: `http://SEU_IP:8097`

### Autenticação

Todas as requisições precisam do header:
```
Authorization: Bearer SUA_API_KEY
```

### Principais Endpoints

```
POST /instance                    # Criar instância
GET  /instance                    # Listar instâncias
POST /instance/:id/connect        # Conectar (QR Code)
POST /instance/:id/logout         # Desconectar
GET  /instance/:id/status         # Status

POST /message/text                # Enviar texto
POST /message/audio               # Enviar áudio
POST /message/image               # Enviar imagem
POST /message/video               # Enviar vídeo
POST /message/document            # Enviar documento
POST /message/missedCall          # Chamada perdida (experimental)

POST /chat/deleteChat             # Deletar chat
POST /chat/archiveChat            # Arquivar chat
POST /chat/read-messages          # Marcar como lido
POST /chat/presence               # Enviar presença
```

## 🔐 Segurança

### Alterar Senhas Padrão

**IMPORTANTE:** Altere as senhas padrão antes de usar em produção:

1. **API_KEY:**
```yaml
- API_KEY=sua_chave_segura_aqui
```

2. **PostgreSQL:**
```yaml
- POSTGRES_PASSWORD=sua_senha_postgres
- DB_URL=postgres://whatsmiau:sua_senha_postgres@postgres_whatsmiau:5432/...
```

3. **Redis:**
```yaml
- REDIS_PASSWORD=sua_senha_redis
# E no serviço redis_whatsmiau:
--requirepass sua_senha_redis
```

## 📈 Monitoramento

### Health Checks

A stack inclui health checks automáticos:

- **API:** `curl http://localhost:8097/health`
- **PostgreSQL:** `pg_isready`
- **Redis:** `redis-cli ping`

### Logs

```bash
# API
docker service logs -f whatsmiau_api

# PostgreSQL
docker service logs -f postgres_whatsmiau

# Redis
docker service logs -f redis_whatsmiau
```

### Métricas

```bash
# Ver uso de recursos
docker stats

# Ver serviços
docker service ls
```

## 🛠️ Troubleshooting

### API não inicia

1. Verificar se volumes foram criados
2. Verificar logs: `docker service logs whatsmiau_api`
3. Verificar se PostgreSQL está pronto: `docker service logs postgres_whatsmiau`

### Erro de conexão com banco

1. Aguardar 30-40s (health check do PostgreSQL)
2. Verificar senha no `DB_URL`
3. Verificar se serviço está rodando: `docker service ls`

### Redis não conecta

1. Verificar senha no `REDIS_PASSWORD`
2. Verificar logs: `docker service logs redis_whatsmiau`
3. Testar conexão: `docker exec -it $(docker ps -q -f name=redis) redis-cli -a SENHA ping`

## 📊 Capacidade e Performance

### Configuração Atual

- **Conexões simultâneas:** 100-300 WhatsApp
- **Mensagens/segundo:** ~500 msg/s
- **Webhooks/minuto:** ~2000 hooks/min
- **Goroutines:** ~350 (otimizado)

### Recursos Totais

- **CPU Total:** 7 CPUs (4 API + 2 PostgreSQL + 1 Redis)
- **RAM Total:** 7.5GB (4GB API + 2GB PostgreSQL + 1.5GB Redis)
- **Disco:** Volumes persistentes para PostgreSQL e Redis

## 🔄 Atualização

Para atualizar a imagem:

```bash
# Pull nova versão
docker pull jondevsouza31/miau-jon:latest

# Atualizar serviço (zero downtime)
docker service update --image jondevsouza31/miau-jon:latest whatsmiau_api
```

## 📚 Documentação Completa

Acesse o arquivo `relatorio-api.html` no repositório para documentação completa com todas as features e otimizações.

## 🤝 Suporte

- **Repository:** https://github.com/jonemp31/miau-jon
- **Docker Hub:** https://hub.docker.com/r/jondevsouza31/miau-jon

## 📄 Licença

Este projeto segue os termos de uso do WhatsApp e da biblioteca Whatsmeow.

---

**Versão:** 2.0 (com otimizações de escalabilidade)  
**Status:** ✅ PRODUCTION READY  
**Data:** Novembro 2025
