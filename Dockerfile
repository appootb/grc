# Dashboard API
FROM golang:1.14-alpine

COPY ./ /go/src/github.com/appootb/grc
WORKDIR /go/src/github.com/appootb/grc/dashboard

RUN CGO_ENABLED=0 go build

# Dashboard UI
FROM node:11

COPY dashboard/views /data/views
WORKDIR /data/views

RUN rm -rf node_modules && npm install && npm run build

# Image
FROM alpine:3.9

# Install nginx
RUN apk update && apk add nginx && mkdir -p /run/nginx/

# API
COPY --from=0 /go/src/github.com/appootb/grc/dashboard/dashboard /data/dashboard/bin/dashboard
# UI
COPY --from=1 /data/views/dist /var/lib/nginx/html

# nginx config
COPY dashboard/etc/nginx/default.conf /etc/nginx/conf.d/default.conf
# Run scripts
COPY dashboard/etc/run.sh /run.sh

# API config file
COPY dashboard/config.yaml /etc/config.yaml

EXPOSE 80

CMD ["sh", "/run.sh"]