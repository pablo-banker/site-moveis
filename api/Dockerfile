FROM golang:1.26.8-alpine@sha256:8ac98ca534ac3f51e1f420a1dd2c15e74c75cfa0f23f3ad27eb5d7236c349a0c AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/api ./cmd/server && CGO_ENABLED=0 go build -trimpath -o /out/migrate ./cmd/migrate
FROM alpine:3.24@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6
RUN apk add --no-cache ca-certificates && addgroup -S forma && adduser -S -G forma forma
USER forma
COPY --from=build /out/api /usr/local/bin/forma-api
COPY --from=build /out/migrate /usr/local/bin/forma-migrate
EXPOSE 8081
ENTRYPOINT ["/usr/local/bin/forma-api"]
