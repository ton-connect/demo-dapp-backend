FROM golang:1.18 AS gobuild
WORKDIR /build-dir
COPY go.mod .
COPY go.sum .
RUN go mod download all
COPY . .
RUN go build -o /tmp/tonproof github.com/tonkeeper/tonproof


FROM golang:1.17 AS tonproof
COPY --from=gobuild /tmp/tonproof /app/tonproof
CMD ["/app/tonproof"]


