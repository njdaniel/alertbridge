.PHONY: build test docker clean

build:
	go build -o alertbridge ./cmd/alertbridge

test:
	go test ./...

docker:
	docker build -t alertbridge .

clean:
	rm -f alertbridge 