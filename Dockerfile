# Environment
FROM golang:1.15 as build-env

RUN mkdir -p /opt/anti_bruteforce
WORKDIR /opt/anti_bruteforce
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -v -o "./bin/antifrod" -ldflags "-X main.release="develop" -X main.buildDate=2021-01-26T18:32:26 -X main.gitHash=e452de9" ./cmd

# Release
FROM alpine:latest
COPY --from=build-env /opt/anti_bruteforce /bin/antifrod
ENTRYPOINT ["/bin/antifrod"]

