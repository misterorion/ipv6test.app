FROM public.ecr.aws/docker/library/golang:latest AS build
WORKDIR /app
COPY go.mod go.sum ./
COPY main.go .
RUN go build -tags lambda.norpc -o main main.go

FROM public.ecr.aws/lambda/provided:al2023
COPY index.tmpl ./index.tmpl
COPY --from=build /app/main ./main
ENTRYPOINT [ "./main" ]