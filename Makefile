.PHONY: build run test clean

build:
	go build -o wblue .

run: build
	./wblue start

test:
	go test ./...

clean:
	rm -f wblue
	rm -rf ~/.wblue/data
