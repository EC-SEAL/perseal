FROM golang:latest AS build-env

WORKDIR /app

# Copy go mod and sum files
COPY go.mod .
COPY go.sum .

# Download all dependancies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

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


FROM tomcat:8.0.47-jre7 AS tomcat
ARG SSL_KEYSTORE_PATH=keystore.jks
ARG SSL_CERT_ALIAS=perseal
ARG SSL_KEY_PASS=keystorepass
ARG INTER=inter.p12

COPY keystore.jks .
COPY public.pub .

RUN keytool -importkeystore -srckeystore ./$SSL_KEYSTORE_PATH -srcstorepass $SSL_KEY_PASS -srcalias $SSL_CERT_ALIAS -destalias $SSL_CERT_ALIAS -destkeystore $INTER -deststoretype PKCS12 -deststorepass $SSL_KEY_PASS
RUN openssl pkcs12 -in $INTER -nodes -nocerts -out private.key -passin pass:$SSL_KEY_PASS
RUN ls .
# Run the executable

RUN mkdir ./ui
COPY ui ui

COPY --from=build-env /go/bin/perseal /go/bin/perseal

#ENTRYPOINT ["/go/bin/perseal"]	