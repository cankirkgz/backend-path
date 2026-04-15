# 1. aşama: build aşaması
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Önce go mod dosyalarını kopyala
COPY go.mod go.sum ./

# Dependency'leri indir
RUN go mod download

# Projedeki tüm dosyaları kopyala
COPY . .

# Uygulamayı derle
RUN CGO_ENABLED=0 GOOS=linux go build -o backend-app ./cmd/api

# 2. aşama: daha küçük final image
FROM alpine:3.20

WORKDIR /app

# Sertifika paketlerini yükle
RUN apk add --no-cache ca-certificates

# Builder aşamasından derlenmiş binary'yi al
COPY --from=builder /app/backend-app .

# Uygulamanın çalışacağı port
EXPOSE 8080

# Uygulamayı başlat
CMD ["./backend-app"]