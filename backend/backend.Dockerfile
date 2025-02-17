# Официальный образ Go
FROM golang:1.24-alpine

# Установка зависимостей для SQLite3 и CGO
RUN apk add --no-cache sqlite-dev gcc musl-dev

# Рабочая директория в контейнере
WORKDIR /app

# Копировать go.mod и go.sum
COPY go.mod go.sum ./

# Загрузить зависимости
RUN go mod download

# Копируем проект
COPY internal ./internal
COPY main.go ./main.go
COPY dist ./dist

# Сборка приложения
# RUN go build -o ./ ./main.go
# Сборка с включенным CGO
RUN CGO_ENABLED=1 go build -o ./ ./main.go

# Запуск приложения
CMD ["./main"]