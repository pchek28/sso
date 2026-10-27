FROM golang:1.26.5-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/cmd/sso/exe /app/cmd/sso

FROM alpine:3.24.1
WORKDIR /app
COPY --from=builder /app/cmd/sso/exe /app
COPY --from=builder /app/web /app/web
ENV PROJECT_ROOT=/app
CMD ["/app/exe"]
