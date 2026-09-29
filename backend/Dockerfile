FROM golang:1.27-alpine AS build
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
ARG SERVICE
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/service ./cmd/${SERVICE}

FROM alpine:3.22
RUN apk add --no-cache ca-certificates && addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=build /out/service /app/service
COPY backend/migrations /app/migrations
COPY content /content
USER app
ENTRYPOINT ["/app/service"]
