ARG GO_VERSION=1.19
FROM golang:1.19.2-alpine as builder

WORKDIR /app

COPY ./go.mod ./go.sum ./
RUN go mod download

#RUN go mod init hello-cicd
COPY . ./
WORKDIR /app/cmd/hellosvc
RUN CGO_ENABLED=0 GOOS=linux go build -o /hello-cicd

FROM gcr.io/distroless/base-debian11 as final
WORKDIR /
COPY --from=builder /hello-cicd /hello-cicd

EXPOSE 8081
EXPOSE 8082

USER nonroot:nonroot
ENTRYPOINT ["/hello-cicd"]