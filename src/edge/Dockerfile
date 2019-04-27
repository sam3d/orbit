FROM golang:latest AS build
WORKDIR /build/edge
COPY . ..
RUN go install
RUN go build -o startup

FROM debian:latest AS certs
WORKDIR /certs
RUN apt-get update && apt-get install --yes openssl
RUN openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 1024 -nodes -subj '/CN=localhost'

FROM nginx:stable
COPY --from=build /build/edge/startup /opt
COPY --from=certs /certs /etc/nginx/certs/dummy
COPY edge/default.conf /etc/nginx/conf.d/default.conf
ENTRYPOINT /opt/startup && nginx -g "daemon off;"
