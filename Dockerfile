FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/goreecloud-music ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/goreecloud-music /usr/local/bin/goreecloud-music
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/goreecloud-music"]
