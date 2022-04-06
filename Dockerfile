####
# Build the web frontend
####

FROM alpine AS builder-web
WORKDIR /web/
RUN apk add --no-cache nodejs npm make git

# Make sure npm is up to date
RUN npm install -g npm

# Install yarn for web dependency management
RUN npm install -g yarn

# Install polymer CLI
RUN yarn global add polymer-cli

# Copy web source files
COPY web/ .

# Build the frontend
RUN make

####
# Build the go binary
####

FROM golang:alpine AS builder-go
RUN apk add --no-cache libjpeg-turbo-dev git g++ make
WORKDIR /go/src/timelapse-queue/

# Copy all source files.
COPY . .

# Copy built web package from the previous stage.
COPY --from=builder-web /web/build/ /go/src/timelapse-queue/web/build/

# Install go-bindata-assetfs executable for constructing our archive.
RUN go get github.com/go-bindata/go-bindata/...
RUN go get github.com/elazarl/go-bindata-assetfs/...

# Build the standalone executable.
RUN make build

####
# Compose everything into the final minimal image.
####

FROM alpine
WORKDIR /app
COPY --from=builder-go /go/src/timelapse-queue/timelapse-queue /app
RUN apk add --no-cache ffmpeg libjpeg-turbo

# Create the mountpoint. The user is expected to run the image with a
# filesystem bound here.
RUN mkdir -p /mnt/fsroot

# Use local timezone.
# TODO use system time instead of hardcoded.
RUN apk add --update tzdata
ENV TZ=America/Los_Angeles

EXPOSE 80
CMD ["./timelapse-queue", "--port", "80", "--root", "/mnt/fsroot"]
