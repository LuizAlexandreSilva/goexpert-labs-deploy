FROM golang:latest as builder
WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build --ldflags="-w -s" -o lab-deploy cmd/server.go

FROM alpine:latest
COPY --from=builder /app/lab-deploy /app/lab-deploy
CMD ["/app/lab-deploy"]
