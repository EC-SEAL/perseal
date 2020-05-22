FROM golang:latest AS build-env

WORKDIR /app

# Copy go mod and sum files
COPY perseal/go.mod .
COPY perseal/go.sum .

# Download all dependancies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY perseal/ .

# Add variables to system environment
ADD .env /tmp/.env
RUN cat /tmp/.env >> /etc/environment

# Build the Go app
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/perseal \
    -ldflags "-extldflags \"-fno-PIC -static \
      -lpthread -lstdc++\"" -buildmode pie -tags 'osusergo netgo static_build'

FROM scratch
ARG PERSEAL_INT_PORT
ARG PERSEAL_EMAIL
EXPOSE $PERSEAL_INT_PORT
LABEL maintainer=$PERSEAL_EMAIL

FROM alpine:latest as alpine
RUN apk add -U --no-cache ca-certificates

FROM tomcat:8.0.47-jre7
ARG KEYSTORE=keystore.jks
ARG ALIAS=perseal
ARG KEYSTOREPASS=keystorepass
ARG INTER=inter.p12

RUN apt-get update && apt-get install -y xdg-utils


COPY perseal/keystore.jks .
COPY perseal/public.pub .
RUN keytool -importkeystore -srckeystore ./$KEYSTORE -srcstorepass $KEYSTOREPASS -srcalias $ALIAS -destalias $ALIAS -destkeystore $INTER -deststoretype PKCS12 -deststorepass $KEYSTOREPASS
RUN openssl pkcs12 -in $INTER -nodes -nocerts -out private.key -passin pass:$KEYSTOREPASS
RUN ls .


COPY --from=build-env /go/bin/perseal /go/bin/perseal
# Run the executable
