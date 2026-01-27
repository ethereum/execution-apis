.PHONY: all clean build buildgeth test lint fill

all: clean build

build:
	go build .

buildgeth:
	go build -o geth github.com/ethereum/go-ethereum/cmd/geth

clean:
	rm -rf rpctestgen tests

test:
	go test ./...

lint:
	gofmt -d ./
	go vet ./...
	staticcheck ./...

fill: build buildgeth
	./rpctestgen --bin ./geth -chain ./chain --out ../tests
