FROM --platform=$BUILDPLATFORM golang:1 AS build
ARG TARGETOS TARGETARCH
WORKDIR /work
COPY . .
RUN go generate && CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o app .

FROM scratch
COPY --from=build /work/app /app
EXPOSE 8080
ENTRYPOINT ["/app"]
