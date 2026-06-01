BINARY := main

build:
	go build -o $(BINARY) ./cmd/...

run: build
	./$(BINARY)

test:
	go test -race ./...

lint:
	go vet ./...

clean:
	rm -f $(BINARY)

.PHONY: build run test lint clean
