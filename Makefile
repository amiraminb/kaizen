APP := kaizen
PKG := ./...
BIN_DIR := ~/.local/bin

.PHONY: build dep test fmt vet clean

build: dep
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP) .

dep:
	go mod tidy
	go mod vendor

test:
	go test $(PKG)

fmt:
	go fmt $(PKG)

vet:
	go vet $(PKG)

clean:
	rm -rf $(BIN_DIR)/$(APP)
