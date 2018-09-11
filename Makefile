all:
	./build.sh

debug:
	./build.sh debug

clean:
	rm -r web/build
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
