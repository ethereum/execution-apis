.PHONY: all clean build buildgeth test lint fill

all: clean build

build:
	go build .

buildgeth:
	go install github.com/ethereum/go-ethereum/cmd/geth@master

clean:
	rm -rf rpctestgen tests

test:
	go test ./...

lint:
	gofmt -d ./
	go vet ./...
	staticcheck ./...

fill: build buildgeth
	./rpctestgen --bin `which geth` -chain ./chain --out ../tests
