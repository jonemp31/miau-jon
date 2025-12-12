# 🔐 Device Fingerprint Persistent

## Visão Geral

A partir da versão **v3.2-fingerprint-persistent**, a Miau-Jon suporta **fingerprints de dispositivos persistentes** para cada instância. O fingerprint é definido **uma única vez na criação da instância** e permanece o mesmo para sempre, simulando um dispositivo físico real.

## 🎯 O que são Fingerprints?

Fingerprints (impressões digitais) são as características técnicas que identificam o tipo de dispositivo conectado ao WhatsApp:
- **Sistema Operacional**: Windows, macOS, Ubuntu
- **Navegador**: Chrome, Firefox, Safari, Edge  
- **Versão do OS**: 10.0, 11.0, 22.04, etc.

**IMPORTANTE**: O fingerprint é definido **uma única vez na criação da instância** e permanece o mesmo para sempre, simulando um dispositivo físico real.

## ✅ Navegadores Suportados

### 1. **Chrome** (Padrão)
- **OS**: Windows 10
- **Plataforma**: CHROME
- **User-Agent**: `WhatsApp/2.2445.9 Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36`
- **Uso**: `fingerprintType: "chrome"` ou omitir o campo

### 2. **Firefox**
- **OS**: Ubuntu 22.04
- **Plataforma**: FIREFOX
- **User-Agent**: `WhatsApp/2.2445.9 Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:121.0) Gecko/20100101 Firefox/121.0`
- **Uso**: `fingerprintType: "firefox"`

### 3. **Safari**
- **OS**: macOS Sonoma 14.5
- **Plataforma**: SAFARI
- **User-Agent**: `WhatsApp/2.2445.9 Mozilla/5.0 (Macintosh; Intel Mac OS X 14_5) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4.1 Safari/605.1.15`
- **Uso**: `fingerprintType: "safari"`

### 4. **Edge**
- **OS**: Windows 11
- **Plataforma**: EDGE
- **User-Agent**: `WhatsApp/2.2445.9 Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.2210.144`
- **Uso**: `fingerprintType: "edge"`

## 📌 Como Funciona

### Ciclo de Vida do Fingerprint

```
1. CRIAÇÃO DA INSTÂNCIA
   ↓
   Escolhe fingerprintType (chrome, firefox, safari, edge)
   ↓
   Salva no banco de dados
   ↓
2. CONEXÃO (QR Code ou Pairing)
   ↓
   Busca fingerprintType do banco automaticamente
   ↓
   Aplica ao cliente WhatsApp
   ↓
3. TODAS AS CONEXÕES FUTURAS
   ↓
   Sempre usa o mesmo fingerprint salvo
```

## 🚀 Como Usar

### 1️⃣ Criar Instância com Fingerprint (API)

```bash
POST /v1/instance/create
Content-Type: application/json
x-api-key: SUA_API_KEY

{
    "instanceName": "Instancia Firefox",
    "webhook": "https://meuwebhook.com/callback",
    "webhookByEvents": true,
    "alwaysOnline": true,
    "readMessages": true,
    "fingerprintType": "firefox"  # ← DEFINE AQUI (chrome, firefox, safari, edge)
}
```

**Resposta**:
```json
{
    "id": "1",
    "instanceName": "Instancia Firefox",
    "fingerprintType": "firefox",  # ← SALVO NO BANCO
    "connected": false
}
```

### 2️⃣ Conectar QR Code (API)

```bash
POST /v1/instance/1/connect
Content-Type: application/json
x-api-key: SUA_API_KEY

{}  # ← NÃO PRECISA ENVIAR fingerprintType (usa o salvo automaticamente)
```

**O sistema automaticamente**:
1. Busca `fingerprintType` do banco (`firefox` neste exemplo)
2. Aplica ao cliente WhatsApp antes de gerar o QR Code
3. Retorna o QR Code com o fingerprint correto

### 3️⃣ Conectar Pairing Code (API)

```bash
POST /v1/instance/1/pair
Content-Type: application/json
x-api-key: SUA_API_KEY

{
    "phoneNumber": "5521999999999"
    # ← NÃO PRECISA ENVIAR fingerprintType (usa o salvo automaticamente)
}
```

**O sistema automaticamente**:
1. Busca `fingerprintType` do banco (`firefox` neste exemplo)
2. Aplica ao cliente WhatsApp antes de solicitar o código
3. Retorna o código de pareamento com o fingerprint correto

### 4️⃣ Usar Interface Web (test-api.html)

1. Acesse `http://localhost:3000/test-api.html`
2. Na seção **"Create Instance"**:
   - Preencha o nome da instância
   - **Selecione o Fingerprint** no dropdown (Chrome, Firefox, Safari, Edge)
   - Configure webhook e outras opções
   - Clique em "Create Instance"
3. Na seção **"QR Code"**:
   - Cole o ID da instância
   - Clique em "Show QR Code"
   - ⚠️ O fingerprint usado será o que você escolheu na criação
4. Na seção **"Pairing Code"**:
   - Cole o ID da instância e número de telefone
   - Clique em "Request Pairing Code"
   - ⚠️ O fingerprint usado será o que você escolheu na criação

## 🔄 Comparação v3.1 vs v3.2

| Aspecto | v3.1 (Incorreto) | v3.2 (Correto) |
|---------|------------------|----------------|
| **Quando escolher** | No QR Code/Pairing | ✅ **Na criação** |
| **Persistência** | ❌ Não persistia | ✅ **Salvo no DB** |
| **Consistência** | ❌ Podia mudar | ✅ **Permanente** |
| **Realismo** | ❌ Suspeito | ✅ **Natural** |
| **Exemplo** | Hoje Chrome, amanhã Firefox | Sempre Firefox |

### Por que v3.1 estava incorreto?

```
❌ v3.1: Instância podia conectar como Chrome hoje e Firefox amanhã
→ WhatsApp pensa: "Por que este dispositivo mudou de sistema operacional?"
→ Padrão suspeito e inconsistente

✅ v3.2: Instância sempre conecta com o mesmo fingerprint
→ WhatsApp pensa: "Este dispositivo é sempre o mesmo"
→ Comportamento natural e realista
```

## 🛡️ Retrocompatibilidade

- **Instâncias criadas antes de v3.2** (sem `fingerprintType` no banco):
  - Usarão automaticamente **Chrome (Windows 10)** por padrão
  - Nenhuma ação necessária
  
- **Instâncias criadas após v3.2**:
  - Usarão o fingerprint escolhido na criação
  - Se não especificado, usarão Chrome por padrão

## ✨ Benefícios

- ✅ **Consistência**: Uma instância = um dispositivo único para sempre
- ✅ **Realismo**: Simula dispositivos físicos reais que não mudam de identidade
- ✅ **Segurança**: Evita padrões suspeitos de mudança de dispositivo
- ✅ **Diversificação**: Cada instância pode ter um fingerprint diferente
- ✅ **Simplicidade**: Escolhe uma vez, usa para sempre
- ✅ **Auditoria**: Fácil rastrear qual dispositivo cada instância representa

## 🔧 Implementação Técnica

### Model (models/instance.go)

```go
type Instance struct {
    ID              string  `json:"id"`
    InstanceName    string  `json:"instanceName"`
    FingerprintType string  `json:"fingerprintType,omitempty"` // ← NOVO CAMPO
    // ... outros campos
}
```

### Controller - Create (controllers/instance.go)

```go
func (s *InstanceController) Create(c echo.Context) error {
    // ... código anterior
    
    instance := models.Instance{
        ID:              id,
        InstanceName:    instanceName,
        FingerprintType: req.FingerprintType, // ← LÊ DO REQUEST
        // ... outros campos
    }
    
    // Se não especificado, usa Chrome
    if instance.FingerprintType == "" {
        instance.FingerprintType = "chrome"
    }
    
    // Salva no banco
    err = s.repo.Save(&instance)
}
```

### Controller - Connect (controllers/instance.go)

```go
func (s *InstanceController) Connect(c echo.Context) error {
    id := c.Param("id")
    
    // Busca instância do banco
    result, err := s.repo.Find(id)
    
    // Pega fingerprint salvo
    fingerprintType := result[0].FingerprintType
    if fingerprintType == "" {
        fingerprintType = "chrome" // retrocompatibilidade
    }
    
    // Conecta com o fingerprint correto
    _, err = s.whatsmiau.Connect(id, fingerprintType) // ← USA O SALVO
}
```

### Core - Apply Fingerprint (lib/whatsmiau/whatsmeow.go)

```go
func applyFingerprintProfile(fingerprintType string) {
    switch fingerprintType {
    case "chrome":
        store.DeviceProps.Os = proto.String("Windows 10")
        store.DeviceProps.PlatformType = waProto.DeviceProps_CHROME.Enum()
        store.DeviceProps.RequireFullSync = proto.Bool(false)
    case "firefox":
        store.DeviceProps.Os = proto.String("Ubuntu 22.04")
        store.DeviceProps.PlatformType = waProto.DeviceProps_FIREFOX.Enum()
        store.DeviceProps.RequireFullSync = proto.Bool(false)
    case "safari":
        store.DeviceProps.Os = proto.String("macOS 14.5")
        store.DeviceProps.PlatformType = waProto.DeviceProps_SAFARI.Enum()
        store.DeviceProps.RequireFullSync = proto.Bool(false)
    case "edge":
        store.DeviceProps.Os = proto.String("Windows 11")
        store.DeviceProps.PlatformType = waProto.DeviceProps_EDGE.Enum()
        store.DeviceProps.RequireFullSync = proto.Bool(false)
    default:
        // Chrome por padrão
        store.DeviceProps.Os = proto.String("Windows 10")
        store.DeviceProps.PlatformType = waProto.DeviceProps_CHROME.Enum()
        store.DeviceProps.RequireFullSync = proto.Bool(false)
    }
}
```

## 📊 Exemplos de Uso

### Cenário 1: Criar 10 instâncias com fingerprints diferentes

```bash
# Instância 1 - Chrome
curl -X POST http://localhost:3000/v1/instance/create \
  -H "x-api-key: SUA_KEY" \
  -H "Content-Type: application/json" \
  -d '{"instanceName": "Cliente 1", "fingerprintType": "chrome"}'

# Instância 2 - Firefox
curl -X POST http://localhost:3000/v1/instance/create \
  -H "x-api-key: SUA_KEY" \
  -H "Content-Type: application/json" \
  -d '{"instanceName": "Cliente 2", "fingerprintType": "firefox"}'

# Instância 3 - Safari
curl -X POST http://localhost:3000/v1/instance/create \
  -H "x-api-key: SUA_KEY" \
  -H "Content-Type: application/json" \
  -d '{"instanceName": "Cliente 3", "fingerprintType": "safari"}'

# Instância 4 - Edge
curl -X POST http://localhost:3000/v1/instance/create \
  -H "x-api-key: SUA_KEY" \
  -H "Content-Type: application/json" \
  -d '{"instanceName": "Cliente 4", "fingerprintType": "edge"}'
```

### Cenário 2: Conectar instância (usa fingerprint salvo)

```bash
# Conectar instância 2 (Firefox) - NÃO PRECISA ENVIAR FINGERPRINT
curl -X POST http://localhost:3000/v1/instance/2/connect \
  -H "x-api-key: SUA_KEY" \
  -H "Content-Type: application/json" \
  -d '{}'

# Sistema automaticamente:
# 1. Busca fingerprint do banco → "firefox"
# 2. Aplica Firefox ao cliente
# 3. Gera QR Code com Firefox
```

### Cenário 3: Verificar fingerprint de uma instância

```bash
# Buscar detalhes da instância
curl -X GET http://localhost:3000/v1/instance/list \
  -H "x-api-key: SUA_KEY"

# Resposta:
{
  "instances": [
    {
      "id": "1",
      "instanceName": "Cliente 1",
      "fingerprintType": "chrome",  # ← AQUI ESTÁ O FINGERPRINT
      "connected": true
    },
    {
      "id": "2",
      "instanceName": "Cliente 2",
      "fingerprintType": "firefox",  # ← ESTE É FIREFOX
      "connected": false
    }
  ]
}
```

## 🐛 Troubleshooting

### Problema: Instância antiga não tem fingerprint

**Solução**: Instâncias criadas antes de v3.2 usarão automaticamente Chrome por padrão. Se quiser mudar, recrie a instância.

### Problema: Quero mudar o fingerprint de uma instância existente

**Resposta**: Não é possível (por design). O fingerprint é permanente para simular um dispositivo físico real. Se precisar de outro fingerprint, crie uma nova instância.

### Problema: Erro "invalid fingerprint type"

**Solução**: Use apenas: `chrome`, `firefox`, `safari` ou `edge` (minúsculas).

## 📚 Referências

- **WhatsApp Business API**: Device Props
- **Biblioteca**: go.mau.fi/whatsmeow
- **Versão**: v3.2-fingerprint-persistent
- **Última Atualização**: Dezembro 2024

---

💡 **Dica**: Distribua seus clientes entre os 4 fingerprints disponíveis para máxima diversificação!
