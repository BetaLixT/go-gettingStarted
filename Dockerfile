FROM golang:1.19-alpine AS build
WORKDIR /
COPY . .

RUN go build -o /app /cmd/server/main.go

FROM alpine

ENV GIN_MODE release

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app /app
ENTRYPOINT [ "/app" ]
