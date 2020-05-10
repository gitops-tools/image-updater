FROM golang:latest AS build
WORKDIR /go/src
COPY . /go/src
RUN go build ./cmd/quay-hooks

FROM registry.access.redhat.com/ubi8/ubi-minimal
WORKDIR /root/
COPY --from=build /go/src/quay-hooks .
EXPOSE 8080
CMD ["./quay-hooks", "http"]
