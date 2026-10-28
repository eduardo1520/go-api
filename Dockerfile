# --- Stage 1: Build the application ---
FROM golang:alpine AS builder

WORKDIR /go/src/app

# Instala dependências de build e atualiza pacotes para evitar vulnerabilidades no stage de build
RUN apk update && apk upgrade --no-cache && \
    apk add --no-cache git

# Copia os arquivos de dependência primeiro para aproveitar o cache de camadas do Docker
COPY go.mod go.sum ./
RUN go mod download

# Copia o restante do código-fonte
COPY . .

# Compila um binário estático (CGO_ENABLED=0 é a melhor prática para Alpine)
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/main.go

# --- Stage 2: Secure runtime environment ---
FROM alpine:3.20

WORKDIR /go/src/app

# Atualiza todos os pacotes do Alpine para corrigir vulnerabilidades conhecidas (CVEs)
# e instala os certificados CA necessários para requisições HTTPS/TLS
RUN apk update && apk upgrade --no-cache && \
    apk add --no-cache ca-certificates

# Copia apenas o binário compilado do estágio anterior
COPY --from=builder /go/src/app/main .

EXPOSE 8000

CMD ["./main"]