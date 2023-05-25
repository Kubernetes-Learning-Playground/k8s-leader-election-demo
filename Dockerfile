FROM golang:1.18.7-alpine3.15 as builder

WORKDIR /app

# copy modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on

# cache modules
RUN go mod download

# copy source code
COPY main.go main.go
COPY pkg/signals/ signals/
# build
RUN CGO_ENABLED=0 go build \
    -a -o leaderelection main.go

FROM alpine:3.13

RUN apk --no-cache add ca-certificates

USER nobody

COPY --from=builder --chown=nobody:nobody /app/leaderelection .

ENTRYPOINT ["./leaderelection"]