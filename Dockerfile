FROM golang:alpine AS builder

WORKDIR /src/app

RUN apk add --no-cache git curl

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go mod tidy

RUN go build -o main .

FROM alpine:latest

WORKDIR /src/app

COPY --from=builder /src/app/main .

EXPOSE 8080

CMD ["./main"]
