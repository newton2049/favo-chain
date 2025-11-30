FROM golang:1.18-alpine AS builder

WORKDIR /favo-chain

ADD go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o favo-chain main.go

FROM alpine:latest AS runner

RUN apk --no-cache add ca-certificates jq

WORKDIR /favo-chain

COPY --from=builder /favo-chain/favo-chain ./
COPY ./docker/local/favo-chain.sh ./
COPY ./core-contracts/artifacts ./core-contracts/artifacts

# Expose json-rpc, libp2p and grpc ports
EXPOSE 8545 9632 1478 5001

ENTRYPOINT ["./favo-chain.sh"]
