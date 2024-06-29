FROM golang:latest AS builder
WORKDIR /app
COPY . .
RUN go generate && CGO_ENABLED=0 go build -o app .

FROM scratch
COPY --from=builder /app/app /app
EXPOSE 8080
ENTRYPOINT ["/app"]
