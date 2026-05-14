FROM --platform=$BUILDPLATFORM public.ecr.aws/docker/library/golang:latest AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY src/ ./src/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
    go build -ldflags="-s -w" -tags lambda.norpc -o /out/main ./src/main.go

FROM scratch
COPY --from=builder --chmod=755 /out/main ./main
COPY version src/index.tmpl ./
ENTRYPOINT [ "./main" ]
