.PHONY: build run clean

all: init build

init:
	mkdir -p bin/

build: init
	go build -ldflags "-w -s" -o bin/koala main.go

clean:
	rm -rf bin/

run: build
	./bin/koala
