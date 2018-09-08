all:
	./build.sh

debug:
	./build.sh debug

clean:
	rm -r web/build
	rm bindata.go timelapse-queue
