FROM docker.io/library/golang:1.26-alpine AS builder
WORKDIR /project
COPY . .
RUN go get ./...
RUN go run ./cmd/sarfya-generate-emphasis/
RUN go run ./cmd/sarfya-generate-json/
RUN go build ./cmd/sarfya-prod-server

FROM docker.io/library/alpine:latest
WORKDIR /project
COPY --from=builder /project/sarfya-prod-server /root/.fwew/dictionary-v2.txt /project/stress-data.json /project/data-compiled.json ./
CMD ./sarfya-prod-server
