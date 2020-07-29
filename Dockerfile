FROM golang:latest AS build
WORKDIR /go/src
COPY . /go/src
RUN go build ./cmd/image-updater

FROM registry.access.redhat.com/ubi8/ubi-minimal
WORKDIR /root/
COPY --from=build /go/src/image-updater .
EXPOSE 8080
ENTRYPOINT ["/root/image-updater"]
