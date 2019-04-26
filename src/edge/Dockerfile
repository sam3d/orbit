FROM golang:latest as build
WORKDIR /build/edge
COPY . ..
RUN go install
RUN go build -o startup

FROM nginx:stable
COPY --from=build /build/edge/startup /opt
ENTRYPOINT /opt/startup && nginx -g "daemon off;"
