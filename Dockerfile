FROM golang:1.20 AS gobuild
WORKDIR /build-dir
COPY go.mod .
COPY go.sum .
RUN go mod download all
COPY . .
RUN go build -o /tmp/tonproof github.com/tonkeeper/tonproof


FROM ubuntu AS tonproof
RUN apt-get update && \
    apt-get install -y openssl ca-certificates
COPY --from=gobuild /tmp/tonproof /app/tonproof
CMD ["/app/tonproof"]


