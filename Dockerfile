FROM golang:1 as builder

WORKDIR /go/src/github.com/myanimestream/arias/
COPY . .

RUN go get ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o arias ./cmd/arias


FROM alpine:latest
RUN apk --no-cache add ca-certificates aria2

WORKDIR /root/
COPY --from=builder /go/src/github.com/myanimestream/arias/arias .
COPY aria2.conf .aria2/aria2.conf

CMD ["./arias"]