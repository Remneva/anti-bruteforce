# Environment
FROM golang:1.15 as build-env

RUN mkdir -p /opt/anti_bruteforce
WORKDIR /opt/anti_bruteforce
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY .. .
RUN find
RUN CGO_ENABLED=0 go build -o /opt/anti_bruteforce/cmd

# Release
FROM alpine:latest
COPY --from=build-env /opt/service/anti_bruteforce /bin/anti_bruteforce
ENTRYPOINT ["/bin/anti_bruteforce_service"]

