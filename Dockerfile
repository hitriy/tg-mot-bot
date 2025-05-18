# syntax=docker/dockerfile:1
FROM golang:1.24
COPY . /src
WORKDIR /src
RUN go mod tidy
RUN go build cmd/bot/main.go

FROM debian:bookworm
COPY --from=0 /src/main /bin/
RUN apt-get update && apt-get install -y ca-certificates
RUN update-ca-certificates
CMD ["/bin/main"]