# =================== STAGE 1: BUILD ===================FROM golang:1.25-alpine AS builder

FROM golang:1.24.3-alpine AS builder

WORKDIR /app

# Instalar dependências de build

RUN apk add --no-cache git gcc musl-dev sqlite-dev# Install gcc and SQLite dev libraries

RUN apk add build-base sqlite-dev gcc musl-dev

WORKDIR /build

COPY go.mod go.sum ./

# Copiar arquivos de dependências primeiro (melhor cache)RUN go mod download

COPY go.mod go.sum ./

RUN go mod downloadCOPY . .



# Copiar código fonte# Enable CGO

COPY . .RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o whatsmiau main.go



# Build da aplicação com otimizaçõesFROM alpine:latest

RUN CGO_ENABLED=1 GOOS=linux go build \

    -ldflags="-w -s" \RUN apk update && apk add --no-cache ffmpeg mailcap

    -a -installsuffix cgo \

    -o whatsmiau \WORKDIR /app

    .

COPY --from=builder /app/whatsmiau /app/whatsmiau

# =================== STAGE 2: RUNTIME ===================

FROM alpine:latestRUN mkdir /app/data && chmod 777 -R /app/data



# Instalar dependências de runtimeEXPOSE 8081

RUN apk add --no-cache \

    ca-certificates \ENTRYPOINT ["./whatsmiau"]
    tzdata \
    curl \
    ffmpeg \
    mailcap \
    && rm -rf /var/cache/apk/*

# Criar usuário não-root para segurança
RUN addgroup -g 1000 whatsmiau && \
    adduser -D -u 1000 -G whatsmiau whatsmiau

WORKDIR /app

# Copiar binário do stage de build
COPY --from=builder /build/whatsmiau /app/whatsmiau

# Criar diretório de dados
RUN mkdir -p /app/data && chown -R whatsmiau:whatsmiau /app

# Trocar para usuário não-root
USER whatsmiau

# Expor porta
EXPOSE 8097

# Health check
HEALTHCHECK --interval=30s --timeout=5s --retries=3 --start-period=10s \
    CMD curl -f http://localhost:8097/health || exit 1

# Comando de execução
CMD ["/app/whatsmiau"]
