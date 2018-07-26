all:
	./build.sh

clean:
	rm -r web/build
	rm bindata.go timelapse-queue
