FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o taskflow ./cmd/server

FROM scratch
COPY --from=builder /app/taskflow /taskflow
EXPOSE 8080
CMD ["/taskflow"]
