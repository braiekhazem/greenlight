# Build the Go Binary.
FROM golang:1.22.2 AS builder
ENV CGO_ENABLED=0
ARG BUILD_REF

RUN mkdir -p /build
WORKDIR /build
COPY go.mod /build/go.mod
COPY go.sum /build/go.sum
RUN go mod download

COPY cmd /build/cmd
COPY pkg /build/pkg

# Build the service binary.
RUN go build  -ldflags "-X main.build=${BUILD_REF}"  -o /tmp/bbb-recording-upload ./cmd/main.go


# Run the Go Binary in Alpine.
FROM gcr.io/distroless/static:nonroot
ARG BUILD_DATE
ARG BUILD_REF

COPY --from=builder  /tmp/bbb-recording-upload /bbb-recording-upload

USER root
RUN mkdir -p /recordings && chown 65532:65532 /recordings

USER 65532:65532


CMD ["/bbb-recording-upload"]
