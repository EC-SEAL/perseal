FROM golang:1.13.6-buster AS build-env

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

COPY --from=build-env /go/bin/perseal /go/bin/perseal
# Run the executable

#ENTRYPOINT ["/go/bin/perseal"]
