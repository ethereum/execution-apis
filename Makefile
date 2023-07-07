all: clean build

build:
	go build . 

clean:
	rm -rf rpctestgen tests

test:
	go test ./...

lint:
	gofmt -d ./
	go vet ./...
	staticcheck ./...

fill: build
	./rpctestgen
