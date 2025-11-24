FROM golang:1.22 AS builder
WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/server ./cmd/server

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=builder /out/server /app/server

# No hardcoded ports; read APP_HOST/APP_PORT from environment
USER nonroot:nonroot
ENTRYPOINT ["/app/server"]


