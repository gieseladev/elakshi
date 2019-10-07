FROM golang:1 as builder

WORKDIR /build/elakshi

COPY go.mod ./
COPY go.sum ./

RUN go mod download
RUN go mod verify

COPY cmd ./cmd
COPY pkg ./pkg

RUN CGO_ENABLED=0 GOOS=linux \
    go build -a -installsuffix cgo \
    -o elakshid \
    ./cmd/elakshid


FROM scratch
LABEL maintainer="team@giesela.dev"

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/elakshi/elakshid ./

ENTRYPOINT ["./elakshid"]