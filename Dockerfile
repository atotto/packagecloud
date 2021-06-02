FROM golang:1.16 as builder
WORKDIR /workspace
COPY go.* ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -v -o packagecloud

FROM alpine:latest
RUN apk add --no-cache ca-certificates bash
COPY --from=builder /workspace/packagecloud /bin/packagecloud
