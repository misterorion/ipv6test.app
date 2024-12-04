FROM public.ecr.aws/docker/library/golang:latest AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -tags lambda.norpc -o main main.go

FROM scratch
COPY --from=build /app/main ./main
COPY version index.tmpl ./
ENTRYPOINT [ "./main" ]