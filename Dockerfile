FROM golang AS build

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 go install -v .

FROM gcr.io/distroless/static

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /go/bin/http-server /usr/local/bin/http-server

ENTRYPOINT ["/usr/local/bin/http-server"]