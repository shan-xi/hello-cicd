ARG GO_VERSION=1.19
FROM golang:1.19.2-alpine as builder

WORKDIR /app

#COPY ./go.mod ./go.sum ./
#RUN go mod download

COPY . ./
#RUN go mod init hello-cicd

#WORKDIR /app/cmd/hellosvc
RUN CGO_ENABLED=0 GOOS=linux go build /app/cmd/hellosvc/hellosvc.go

FROM gcr.io/distroless/base-debian11 as final
WORKDIR /
COPY --from=builder /hellosvc /hellosvc

EXPOSE 8081
EXPOSE 8082

USER nonroot:nonroot
ENTRYPOINT ["/hellosvc"]