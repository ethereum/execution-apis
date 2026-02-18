.PHONY: all clean build buildgeth test lint fill

all: build fill

build:
	go build .

buildgeth:
	go build -o geth github.com/ethereum/go-ethereum/cmd/geth

clean:
	rm -f rpctestgen

test:
	go test ./...

lint:
	gofmt -d ./
	go vet ./...
	staticcheck ./...

fill: build buildgeth
	./rpctestgen --bin ./geth -chain ./chain --out ../tests
