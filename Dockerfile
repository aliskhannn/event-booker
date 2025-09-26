FROM golang:alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o event-booker ./cmd/event-booker/main.go

EXPOSE 8080