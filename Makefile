ts := $(shell /bin/date "+%s")

all: check build

check:
	./checkdeps.sh

web/build:
	cd web && $(MAKE)

assetfs: web/build
	go-bindata-assetfs web/build/default/...

debugfs: web/build
	go-bindata-assetfs -debug web/build/default/...

go:
	go mod download
	go get -d -u -v  # Attempt upgrade
	go build -ldflags "-X main.BuildTimestamp=$(ts)"

.PHONY: debug build

debug: debugfs go
build: assetfs go


clean:
	cd web && $(MAKE) clean
	rm bindata.go timelapse-queue


install:
	mkdir -p /usr/local/bin/timelapse/
	cp timelapse-queue /usr/local/bin/timelapse/
	cp timelapse.service /usr/local/bin/timelapse/
	systemctl link /usr/local/bin/timelapse/timelapse.service
 
stop:
	systemctl stop timelapse.service

start:
	systemctl start timelapse.service

reinstall: stop install start
