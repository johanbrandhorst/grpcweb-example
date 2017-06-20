# Build stage
FROM golang AS build-env
ADD . /go/src/github.com/johanbrandhorst/grpcweb-example
ENV CGO_ENABLED=0
RUN cd /go/src/github.com/johanbrandhorst/grpcweb-example && go build -o /app

# Production stage
# Auto-LetsEncrypt requires ca-certificates
FROM broady/cacerts
COPY --from=build-env /app /

# Cache LetsEncrypt certificates
VOLUME /certs

EXPOSE 443
EXPOSE 80
ENTRYPOINT ["/app"]
