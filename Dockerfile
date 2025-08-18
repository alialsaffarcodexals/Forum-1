# ---- build ----
FROM golang:1.22 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download || true
COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /out/server ./cmd/server

# ---- final ----
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/server ./server
COPY web ./web
COPY internal/db/schema.sql ./internal/db/schema.sql
ENV PORT=8080
ENV DB_PATH=/app/data/forum.db
EXPOSE 8080
ENTRYPOINT ["./server"]
