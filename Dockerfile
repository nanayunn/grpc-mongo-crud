FROM golang:1.15-alpine
RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh g++
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY * ./
RUN go mod tidy && make
CMD ["/app/bin/server"]
