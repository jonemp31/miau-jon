# 🔐 Device Fingerprint Rotation

## Visão Geral

A partir da versão v3.1-fingerprint, a Miau-Jon suporta **rotação de fingerprints de dispositivos** para aumentar a diversidade das conexões e reduzir padrões detectáveis pelo WhatsApp.

## 🎯 O que são Fingerprints?

Fingerprints (impressões digitais) são as características técnicas que identificam o tipo de dispositivo conectado ao WhatsApp:
- **Sistema Operacional**: Windows, macOS, Ubuntu
- **Navegador**: Chrome, Firefox, Safari, Edge  
- **Versão do OS**: 10.0, 11.0, 22.04, etc.

## ✅ Navegadores Suportados

### 1. **Chrome** (Padrão)
- **OS**: Windows 10
- **Plataforma**: CHROME
- **Uso**: `fingerprintType: "chrome"` ou omitir o campo

### 2. **Firefox**
- **OS**: Ubuntu 22.04
- **Plataforma**: FIREFOX
- **Uso**: `fingerprintType: "firefox"`

### 3. **Safari**
- **OS**: macOS Sonoma 14.5
- **Plataforma**: SAFARI
- **Uso**: `fingerprintType: "safari"`

### 4. **Edge**
- **OS**: Windows 11
- **Plataforma**: EDGE
- **Uso**: `fingerprintType: "edge"`

## 📡 Como Usar

### Via API REST

**Endpoint**: `POST /v1/instance/:id/connect`

**Request Body**:
```json
{
  "fingerprintType": "firefox"
}
```

**Exemplos**:

```bash
# Conectar com Chrome (padrão)
curl -X POST http://localhost:8097/v1/instance/minha-instancia/connect \
  -H "apikey: test123" \
  -H "Content-Type: application/json"

# Conectar com Firefox
curl -X POST http://localhost:8097/v1/instance/minha-instancia/connect \
  -H "apikey: test123" \
  -H "Content-Type: application/json" \
  -d '{"fingerprintType": "firefox"}'

# Conectar com Safari
curl -X POST http://localhost:8097/v1/instance/outra-instancia/connect \
  -H "apikey: test123" \
  -H "Content-Type: application/json" \
  -d '{"fingerprintType": "safari"}'

# Conectar com Edge
curl -X POST http://localhost:8097/v1/instance/terceira-instancia/connect \
  -H "apikey: test123" \
  -H "Content-Type: application/json" \
  -d '{"fingerprintType": "edge"}'
```

### Via PowerShell

```powershell
# Chrome (padrão)
$headers = @{
    'apikey' = 'test123'
    'Content-Type' = 'application/json'
}
Invoke-RestMethod -Uri 'http://localhost:8097/v1/instance/minha-instancia/connect' `
    -Method POST -Headers $headers

# Firefox
$body = '{"fingerprintType": "firefox"}'
Invoke-RestMethod -Uri 'http://localhost:8097/v1/instance/minha-instancia/connect' `
    -Method POST -Headers $headers -Body $body

# Safari
$body = '{"fingerprintType": "safari"}'
Invoke-RestMethod -Uri 'http://localhost:8097/v1/instance/outra-instancia/connect' `
    -Method POST -Headers $headers -Body $body
```

## 🔄 Estratégias de Rotação

### 1. **Rotação por Instância**
Cada instância usa um fingerprint diferente:
```
instancia-1 → Chrome (Windows 10)
instancia-2 → Firefox (Ubuntu 22.04)
instancia-3 → Safari (macOS 14.5)
instancia-4 → Edge (Windows 11)
instancia-5 → Chrome (Windows 10)  # Ciclo reinicia
```

### 2. **Rotação Aleatória**
Escolha aleatoriamente para cada nova conexão:
```javascript
const fingerprints = ['chrome', 'firefox', 'safari', 'edge'];
const random = fingerprints[Math.floor(Math.random() * fingerprints.length)];
```

### 3. **Rotação por Horário**
Diferentes fingerprints em diferentes períodos:
```
00:00-06:00 → Safari
06:00-12:00 → Chrome
12:00-18:00 → Firefox
18:00-24:00 → Edge
```

## ⚠️ Limitações Importantes

### **Race Condition em Alta Concorrência**
O `DeviceProps` é uma variável **global** da biblioteca whatsmeow. Se duas instâncias conectarem simultaneamente:
```
Request A (firefox) → Define DeviceProps = Firefox
Request B (safari)  → Define DeviceProps = Safari  
Request A conecta  → Pode pegar Safari em vez de Firefox ❌
```

**Solução**: Evite conectar múltiplas instâncias no mesmo segundo. Use delays:
```bash
# Correto
curl -X POST .../instancia-1/connect -d '{"fingerprintType":"chrome"}'
sleep 2
curl -X POST .../instancia-2/connect -d '{"fingerprintType":"firefox"}'
```

### **Android/iOS não suportados**
Apenas navegadores web são seguros. Simular Android/iOS como companion é detectado pelo WhatsApp e pode causar ban.

## 🎯 Melhores Práticas

✅ **Faça**: 
- Rotacionar entre os 4 navegadores disponíveis
- Usar delays de 1-2 segundos entre conexões
- Combinar com IPs/proxies diferentes
- Documentar qual fingerprint cada instância usa

❌ **Não Faça**:
- Conectar 10+ instâncias simultaneamente
- Mudar fingerprint de uma instância já conectada (desconecte antes)
- Usar sempre o mesmo fingerprint para todas as instâncias
- Tentar simular Android/iOS (não suportado)

## 📊 Monitoramento

Os logs mostrarão qual fingerprint foi aplicado:
```
INFO: applied Chrome fingerprint    os="Windows 10"
INFO: applied Firefox fingerprint   os="Ubuntu 22.04"
INFO: applied Safari fingerprint    os="macOS Sonoma 14.5"
INFO: applied Edge fingerprint      os="Windows 11"
```

## 🔧 Troubleshooting

**Problema**: Instância não conecta após mudar fingerprint
- **Solução**: Desconecte completamente antes de reconectar com novo fingerprint

**Problema**: Todas as instâncias aparecem com o mesmo fingerprint
- **Solução**: Adicione delay de 2s entre as conexões

**Problema**: WhatsApp detecta/bane a conta
- **Solução**: Use apenas os 4 navegadores suportados (Chrome/Firefox/Safari/Edge)

## 🚀 Upgrade da v3.0 para v3.1

Nenhuma alteração no banco de dados é necessária. A feature é 100% retrocompatível:
- Se `fingerprintType` não for enviado → usa Chrome (comportamento anterior)
- Se `fingerprintType` for enviado → aplica o fingerprint escolhido

## 📝 Changelog

**v3.1-fingerprint** (12/12/2024)
- ✅ Suporte a 4 fingerprints web (Chrome, Firefox, Safari, Edge)
- ✅ Rotação automática de OS version (Windows 10/11, Ubuntu 22.04, macOS 14.5)
- ✅ API retrocompatível (Chrome default se omitido)
- ✅ Logs detalhados de fingerprint aplicado

---

**Dúvidas?** Entre em contato via Issues no GitHub.
