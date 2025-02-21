# Stage 1: Сборка приложения
FROM golang:1.23-alpine AS builder

RUN apk update && apk upgrade && apk add pkgconf git bash build-base sudo

RUN git clone https://github.com/edenhill/librdkafka.git && cd librdkafka && ./configure --prefix /usr && make && make install


WORKDIR /app

# Устанавливаем зависимости
COPY go.mod go.sum ./
RUN go mod download

ARG NAME="urleater"

# Копируем исходный код и HTML-файлы
COPY . .

# Собираем приложение

RUN go build -tags musl --ldflags "-extldflags -static" -o bin/${NAME} cmd/*.go


# Stage 2: Финальный образ
FROM alpine:latest

RUN apk add --no-cache gcc
# Копируем бинарник из стадии сборки
COPY --from=builder /app/bin/${NAME} .

# Копируем HTML-файлы (например, они лежат в папке "static")
COPY --from=builder /app/templates ./templates

COPY --from=builder /app/build/migrations ./migrations
COPY --from=builder /app/entrypoint.sh ./entrypoint.sh

RUN chmod +x entrypoint.sh

COPY --from=builder /app/build/local/docker.env .env
