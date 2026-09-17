FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/gemini-api .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/gemini-api /app/gemini-api
COPY start.sh /app/start.sh
RUN chmod +x /app/start.sh
ENV PORT=8080 \
    HOST=0.0.0.0 \
    DEFAULT_MODEL=gemini-3.6-flash \
    API_KEYS= \
    IMPERSONATE=
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s \
  CMD wget -qO- http://127.0.0.1:${PORT}/v1/models >/dev/null 2>&1 || exit 1
CMD ["sh", "/app/start.sh"]
