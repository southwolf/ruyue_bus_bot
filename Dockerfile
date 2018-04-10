FROM alpine

RUN apk add --no-cache ca-certificates apache2-utils

ADD ruyue /

CMD ["/ruyue"]