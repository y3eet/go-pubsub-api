# syntax=docker/dockerfile:1

# ---- build stage ----
FROM golang:1.26.1-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build \
  -trimpath \
  -ldflags="-s -w" \
  -o /out/go-pubsub-api \
  ./cmd/api


# ---- runtime stage ----
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

COPY --from=builder /out/go-pubsub-api /go-pubsub-api
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8080

ENV PORT=8080

ENTRYPOINT ["/go-pubsub-api"]