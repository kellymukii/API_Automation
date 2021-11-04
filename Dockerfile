FROM golang:alpine as builder
LABEL maintainer="dmmuriithi"
#ENV http_proxy "http://172.28.200.101:8080"
#ENV https_proxy "http://172.28.200.101:8080"
ENV NO_PROXY="*.safaricom.net,.safaricom.net"
WORKDIR /build
COPY . ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main .
FROM alpine
WORKDIR /build
COPY --from=builder /build/main /build/
COPY configs.yaml configs.yaml
CMD ["./main -c ${CONFIG} -s ${SCENARIOS}"]