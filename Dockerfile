FROM alpine:3.11

COPY app /bin/app

RUN apk --no-cache add curl
RUN apk --no-cache add bind-tools
