FROM golang:1 as builder

# Get the packages manually to alleviate caching
RUN go get\
 cloud.google.com/go/storage\
 github.com/aws/aws-sdk-go/service/s3/s3manager\
 github.com/cenkalti/rpc2\
 github.com/go-chi/chi\
 github.com/google/uuid\
 github.com/gorilla/schema\
 github.com/gorilla/websocket\
 github.com/micro/go-config

WORKDIR /go/src/github.com/myanimestream/arias/
# yes this is stupid, but because of Go's questionable "put everything in the root folder" policy
# there's no other way to do this without rebuilding it every time even when unrelated files change.
COPY aria2/ aria2/
COPY cmd/ cmd/
COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o arias ./cmd/arias


FROM alpine:latest
RUN apk --no-cache add ca-certificates aria2

WORKDIR /root/
COPY --from=builder /go/src/github.com/myanimestream/arias/arias .
COPY .docker/start.sh .

RUN chmod +x arias start.sh

COPY .conf/ /conf/

CMD ["/root/start.sh"]