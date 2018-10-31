# TODO: implementation is work in progress.
# TODO: make separate builder for the frontend instead of relying on locally build data.

####
# Build the go binary
####

FROM golang:alpine AS builder
RUN apk add --no-cache libjpeg-turbo-dev git g++
WORKDIR /go/src/timelapse-queue/

# TODO only copy the relevant part of the repo.
COPY . .

# TODO the go-bindata step from frontend data.

# Fetch all dependencies.
RUN go get -d -v

# Build the main executable.
RUN go build -ldflags "-X main.BuildTimestamp=$(date +%s)"


####
# Compose everything into the final image.
####

FROM alpine
WORKDIR /app
COPY --from=builder /go/src/timelapse-queue/timelapse-queue /app
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
