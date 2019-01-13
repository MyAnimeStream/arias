FROM golang:1 as builder

# Get the packages manually to alleviate caching
RUN go get cloud.google.com/go/storage\
 github.com/cenkalti/rpc2\
 github.com/go-chi/chi\
 github.com/gorilla/schema\
 github.com/gorilla/websocket\
 github.com/micro/go-config

WORKDIR /go/src/github.com/myanimestream/arias/
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o arias ./cmd/arias


FROM alpine:latest
RUN apk --no-cache add ca-certificates aria2

WORKDIR /root/
COPY --from=builder /go/src/github.com/myanimestream/arias/arias .

COPY aria2.conf .aria2/
RUN aria2c

CMD ["./arias"]