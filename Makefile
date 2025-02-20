.PHONY: build install clean

INSTALL_PATH ?= /usr/local/bin
BINARY_NAME = vme

build:
	@echo "Building vme..."
	@go build -o bin/$(BINARY_NAME) cmd/metadataextractor/main.go

install: build
	@echo "Installing vme to $(INSTALL_PATH)..."
	@sudo cp bin/$(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installation complete! Run 'vme -h' to verify."

clean:
	@echo "Cleaning..."
	@rm -rf bin/
