# Stage 1: Build
FROM golang:1.24.3-alpine AS builder

WORKDIR /app

# Instalar dependências de compilação
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Copiar dependências e fazer download
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fonte
COPY . .

# Compilar aplicação
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo -o whatsmiau .

# Stage 2: Runtime
FROM alpine:latest

# Instalar dependências de runtime
RUN apk update && apk add --no-cache ffmpeg mailcap ca-certificates curl tzdata

# Criar usuário não-root
RUN addgroup -g 1000 whatsmiau && adduser -D -u 1000 -G whatsmiau whatsmiau

WORKDIR /app

# Copiar binário compilado
COPY --from=builder /app/whatsmiau /app/whatsmiau

# Criar diretório de dados e ajustar permissões
RUN mkdir -p /app/data && chown -R whatsmiau:whatsmiau /app

# Usar usuário não-root
USER whatsmiau

# Expor porta
EXPOSE 8097

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 CMD curl -f http://localhost:8097/health || exit 1

# Iniciar aplicação
CMD ["/app/whatsmiau"]
