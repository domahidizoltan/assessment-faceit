FROM golang:1.19.3-alpine3.16 AS builder
# FROM golang:1.19.3-alpine3.16
 ENV CGO_ENABLED=0
WORKDIR /go/src/github.com/faceit
COPY . .
RUN go install ./...
RUN go build -o /go/bin/userservice main.go
# ENTRYPOINT ["/go/bin/userservice"]

FROM scratch
COPY --from=builder /go/bin/* /bin/
ENTRYPOINT ["/bin/userservice"]