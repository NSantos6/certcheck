# Stage 1: build
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o certcheck .

# Stage 2: runtime
FROM alpine:3.19

# ca-certificates is required for SSL checks against public CAs
RUN apk --no-cache add ca-certificates tzdata

# OpenShift runs containers with arbitrary non-root UIDs — this ensures
# the binary is executable by any UID in the root group (GID 0)
COPY --from=builder /app/certcheck /usr/local/bin/certcheck
RUN chmod g+x /usr/local/bin/certcheck && \
    chown root:root /usr/local/bin/certcheck

USER 1001

ENTRYPOINT ["certcheck"]
