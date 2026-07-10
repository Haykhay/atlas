FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY atlas /usr/local/bin/atlas
ENTRYPOINT ["atlas"]
