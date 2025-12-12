# 🐳 Stack Portainer - WhatsMiau API

Stack completa para deploy do WhatsMiau API no Portainer (Docker Windows - Standalone Mode)

## 📋 Componentes da Stack

- **whatsmiau_api**: API principal do WhatsApp (porta 8097)
- **redis_whatsmiau**: Cache e gerenciamento de sessões (porta 6379)

**Banco de Dados**: SQLite (arquivo local no volume `whatsmiau_data`)

## 🚀 Como Usar no Portainer

### 1. Acesse o Portainer
```
http://localhost:9000
```

### 2. Criar Nova Stack
1. Vá em **Stacks** no menu lateral
2. Clique em **+ Add stack**
3. Dê um nome: `whatsmiau-api`
4. Cole o conteúdo do `docker-compose.yml` abaixo
5. Clique em **Deploy the stack**

---

## 📦 Docker Compose Stack

```yaml
version: "3.8"

services:
  whatsmiau_api:
    image: jondevsouza31/miau-jon:latest
    container_name: whatsmiau_api
    restart: always
    ports:
      - "8097:8097"
    networks:
      - network_swarm_public
    environment:
      # API Configuration
      PORT: "8097"
      API_KEY: "1234"
      
      # Database (SQLite)
      DIALECT_DB: sqlite3
      DB_URL: file:/app/data/whatsmiau.db?_foreign_keys=on
      
      # Redis Configuration
      REDIS_URL: redis_whatsmiau:6379
      REDIS_PASSWORD: 31121997digiTal100k
      REDIS_TLS: "false"
      
      # Default Settings
      DEFAULT_WEBHOOK_URL: https://webhook-dev.zapsafe.work/webhook/tese324234234
      DEFAULT_AUTO_READ: "true"
      DEFAULT_READ_MESSAGES: "true"
      
      TZ: America/Sao_Paulo
      
    depends_on:
      - redis_whatsmiau
    
    volumes:
      - whatsmiau_data:/app/data

  redis_whatsmiau:
    image: redis:7-alpine
    container_name: redis_whatsmiau
    restart: always
    command: redis-server --requirepass 31121997digiTal100k --appendonly yes
    volumes:
      - redis_data_whatsmiau:/data
    networks:
      - network_swarm_public

volumes:
  whatsmiau_data:
    driver: local
  redis_data_whatsmiau:
    driver: local

networks:
  network_swarm_public:
    driver: bridge
```

---

## 🔧 Configurações Importantes

### 1. **API Key**
```yaml
API_KEY: "1234"
```
Use esta chave no header `apikey` de todas as requisições.

### 2. **Webhook Global**
```yaml
DEFAULT_WEBHOOK_URL: https://webhook-dev.zapsafe.work/webhook/tese324234234
```
Todas as novas instâncias criadas usarão este webhook por padrão.

### 3. **Banco de Dados**
- **Tipo**: SQLite
- **Arquivo**: `/app/data/whatsmiau.db`
- **Volume**: `whatsmiau_data`

### 4. **Redis**
- **Host**: redis_whatsmiau:6379
- **Senha**: 31121997digiTal100k
- **Persistência**: AOF habilitado

---

## 📊 Endpoints da API

### Health Check
```bash
GET http://localhost:8097/v1/health
Header: apikey: 1234
```

### Criar Instância
```bash
POST http://localhost:8097/v1/instance/create
Header: apikey: 1234
Content-Type: application/json

{
  "instanceName": "teste-001",
  "readMessages": true,
  "alwaysOnline": false
}
```

### Pairing Code (Novo!)
```bash
POST http://localhost:8097/v1/instance/:id/pairing
Header: apikey: 1234
Content-Type: application/json

{
  "phoneNumber": "5521999999999"
}
```

### Status do Pairing
```bash
GET http://localhost:8097/v1/instance/:id/pairing/status
Header: apikey: 1234
```

---

## 🎯 Recursos da API

### ✅ Funcionalidades Principais
- ✅ Criação de instâncias WhatsApp
- ✅ Autenticação por QR Code
- ✅ **Autenticação por Pairing Code** (novo!)
- ✅ Envio de mensagens (texto, imagem, áudio, vídeo, documentos)
- ✅ Envio de contatos e localizações
- ✅ Gerenciamento de chats (arquivar, deletar)
- ✅ Presença (digitando, gravando)
- ✅ Webhooks para eventos
- ✅ Auto-read (marca como lido após 8s)
- ✅ Auto-receipt (2 checks cinza)

### 📱 Pairing Code
O Pairing Code é uma forma alternativa de conectar o WhatsApp sem precisar escanear QR Code:

1. Solicite um código via API
2. Receba um código de 8 caracteres (ex: RYBG-Q3YQ)
3. No WhatsApp: Configurações → Aparelhos Conectados → Conectar com número
4. Digite o código
5. Aguarde a conexão (polling automático)

---

## 📈 Monitoramento

### Logs da API
```bash
docker logs -f whatsmiau_api
```

### Logs do Redis
```bash
docker logs -f redis_whatsmiau
```

### Verificar Health
```powershell
Invoke-RestMethod -Uri 'http://localhost:8097/v1/health' -Headers @{'apikey'='1234'}
```

---

## 🔄 Gerenciamento

### Restart dos Serviços
No Portainer:
1. Vá em **Stacks**
2. Selecione `whatsmiau-api`
3. Clique em **Stop** e depois **Start**

Ou via CLI:
```bash
docker-compose restart
```

### Atualizar para Nova Versão
```bash
docker pull jondevsouza31/miau-jon:latest
docker-compose up -d
```

### Backup do Banco de Dados
```bash
# Copiar arquivo SQLite do container
docker cp whatsmiau_api:/app/data/whatsmiau.db ./backup-whatsmiau.db
```

### Restaurar Backup
```bash
# Restaurar arquivo SQLite para o container
docker cp ./backup-whatsmiau.db whatsmiau_api:/app/data/whatsmiau.db
docker restart whatsmiau_api
```

---

## 🛠️ Troubleshooting

### API não inicia
```bash
# Verificar logs
docker logs whatsmiau_api

# Verificar se Redis está rodando
docker exec redis_whatsmiau redis-cli -a 31121997digiTal100k ping
```

### Erro de conexão com Redis
```bash
# Testar conexão
docker exec redis_whatsmiau redis-cli -a 31121997digiTal100k ping

# Verificar logs
docker logs redis_whatsmiau
```

---

## 📚 Documentação Adicional

- [PAIRING_CODE_GUIDE.md](./PAIRING_CODE_GUIDE.md) - Guia completo do Pairing Code
- [README.md](./README.md) - Documentação principal
- [test-api.html](./test-api.html) - Interface web para testes

---

## 🚀 Deploy Automático

Para facilitar o deploy, você pode usar este script PowerShell:

```powershell
# deploy-stack.ps1
cd 'c:\Users\jonat\OneDrive\Área de Trabalho\miau-jon-10-12\miau-jon'

Write-Host "🐳 Atualizando imagem..." -ForegroundColor Cyan
docker pull jondevsouza31/miau-jon:latest

Write-Host "🚀 Iniciando stack..." -ForegroundColor Cyan
docker-compose up -d

Write-Host "⏳ Aguardando inicialização..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

Write-Host "✅ Verificando saúde..." -ForegroundColor Green
Invoke-RestMethod -Uri 'http://localhost:8097/v1/health' -Headers @{'apikey'='1234'}

Write-Host "`n🎉 Deploy concluído!" -ForegroundColor Green
Write-Host "📊 Acesse: http://localhost:8097" -ForegroundColor Cyan
Write-Host "📖 Documentação: test-api.html" -ForegroundColor Cyan
```

---

## 📞 Suporte

- GitHub: https://github.com/jonemp31/miau-jon
- Docker Hub: https://hub.docker.com/r/jondevsouza31/miau-jon

---

**Versão**: v1.0-pairing  
**Data**: 11/12/2024  
**Autor**: Jonathan  
**Status**: ✅ Produção
