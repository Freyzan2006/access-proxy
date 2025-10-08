FROM golang:tip-trixie

WORKDIR app/

COPY . ./go.*

COPY . .

RUN ["go" "build", "-o", "access-proxy"]

RUN ["./build/access-proxy"]