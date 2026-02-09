.PHONY: clean test lint fill

build:
	go build -o rpctestgen ./cmd/rpctestgen
	go build -o speccheck ./cmd/speccheck

geth:
	go build -o geth github.com/ethereum/go-ethereum/cmd/geth

clean:
	rm -rf ./geth ./speccheck ./rpctestgen ../tests

test:
	go test ./...

lint:
	gofmt -d ./
	go vet ./...
	staticcheck ./...

fill: build geth
	./rpctestgen --bin ./geth -chain ./chain --out ../tests
