FROM golang:1.15.8 as builder

WORKDIR /go/src/

COPY . .
RUN go mod download

ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64
RUN go build -o http-example ./main.go

FROM gcr.io/distroless/static:latest as runner

COPY --from=builder /go/src/http-example ./http-example

ENTRYPOINT ["./http-example"]