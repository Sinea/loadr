FROM golang:alpine as base
RUN apk update && apk add --virtual git && apk add bash

# Build the app
FROM base as builder
ADD . ./app
WORKDIR ./app
ENV GO111MODULE=on
RUN CGO_ENABLED=0 go build ./cmd/app/

# Final image
FROM alpine:3.9
COPY --from=builder /go/app/app .
EXPOSE 8080/tcp
EXPOSE 9090/tcp
ENTRYPOINT ["./app"]