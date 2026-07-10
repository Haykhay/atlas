FROM alpine:3.20
RUN apk add --no-cache ca-certificates
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/atlas /usr/local/bin/atlas
ENTRYPOINT ["atlas"]
