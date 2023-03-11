FROM golang:1.17.5-alpine3.15 as modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

FROM golang:1.17.5-alpine3.15 as builder
COPY --from=modules /go/pkg /go/pkg
# RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
COPY . /app
WORKDIR /app
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -ldflags="-s -w" -o ./bin/fcat ./cmd/main.go

FROM alpine
COPY --from=builder /app/config/ /app/config/
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/bin/fcat /app/fcat
COPY --from=builder /app/docs /app/docs
# RUN apk update && apk add ca-certificates --no-cache
# COPY ./fcat.crt /usr/local/share/ca-certificates/fcat.crt
# RUN cat /usr/local/share/ca-certificates/fcat.crt >> /etc/ssl/certs/ca-certificates.crt
# RUN  update-ca-certificates
WORKDIR /app
ENTRYPOINT ["./fcat"]
CMD ["start"]

