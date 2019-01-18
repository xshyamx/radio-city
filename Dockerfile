FROM golang

ADD *.go /go/src/app/

WORKDIR /go/src/app

RUN go get ./... && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w' -o radio-city

FROM scratch

COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=0 /go/src/app/radio-city /

EXPOSE 8080

CMD ["/radio-city"]
