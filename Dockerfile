FROM golang:1.23.4-alpine AS build

WORKDIR /src

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/ai-cloud ./cmd/main.go

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=build /out/ai-cloud /app/ai-cloud

RUN mkdir -p /app/storage_data

EXPOSE 8080

CMD ["/app/ai-cloud"]
