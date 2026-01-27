FROM golang:1.25-alpine AS builder

WORKDIR /build
ENV CGO_ENABLED=1

RUN apk add --no-cache gcc musl-dev sqlite-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION="n/a"

RUN go build -o ./bin/clearnode -ldflags "-X main.Version=${VERSION}" ./clearnode

FROM alpine

RUN addgroup -g 1001 -S clearnode
RUN adduser -S clearnode -u 1001 -G clearnode

USER clearnode

COPY --from=builder /build/bin /bin

EXPOSE 7824 4242

CMD ["clearnode"]

#  => ERROR [builder 7/7] RUN go build -o ./bin/clearnode . -ldflags "-X main.Version=3.0.0"                   0.2s
# ------
#  > [builder 7/7] RUN go build -o ./bin/clearnode . -ldflags "-X main.Version=3.0.0":
# #15 0.188 no Go files in /build
# #15 0.188 malformed import path "-ldflags": leading dash
# #15 0.188 malformed import path "-X main.Version=3.0.0": leading dash
# ------
# executor failed running [/bin/sh -c go build -o ./bin/clearnode . -ldflags "-X main.Version=${VERSION}"]: exit code: 1