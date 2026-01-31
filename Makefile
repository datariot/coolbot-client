.PHONY: build run test clean

BINARY=coolbot-exporter

build:
	go build -o $(BINARY) ./cmd/coolbot-exporter

run: build
	./$(BINARY)

test:
	go test -v ./...

clean:
	rm -f $(BINARY)

deps:
	go mod tidy

docker:
	docker build -t coolbot-exporter .

.PHONY: lint
lint:
	golangci-lint run
