.PHONY: all deps build clean

all: deps build

deps:
	yarn install

build:
	polymer build --preset es6-unbundled --name default

clean:
	rm -r -f build
