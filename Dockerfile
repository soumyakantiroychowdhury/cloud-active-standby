FROM alpine:3.11

COPY app /bin/app

RUN apk --no-cache add curl
