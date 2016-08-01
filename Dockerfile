FROM alpine:3.4

RUN apk update
RUN apk add ca-certificates

COPY logpull /

CMD ["/logpull"]