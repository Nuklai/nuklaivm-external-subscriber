#build stage
FROM golang:alpine AS builder

RUN apk update && apk add --no-cache libc-dev make bash
RUN apk add --virtual build-dependencies build-base
WORKDIR /go/src/app
# Copy the Go application
COPY . .
COPY ./infra/scripts/startup.sh build/
# Build the Go application
RUN go build -o build/subscriber


#final stage
FROM alpine:latest
RUN apk update && apk add --no-cache bash
RUN addgroup -S nuklai && adduser -S nuklai -G nuklai
COPY --from=builder --chown=nuklai /go/src/app/build /app
USER nuklai
RUN chmod a+x /app/startup.sh
ENTRYPOINT [ "/app/startup.sh" ]
LABEL Name=subscriber
EXPOSE 8080
WORKDIR /app
CMD [ "/app/subscriber" ]
