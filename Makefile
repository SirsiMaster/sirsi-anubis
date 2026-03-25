# Pantheon Modular Build System

VERSION ?= v0.4.0-alpha
BUILD_DIR ?= bin
GO_FLAGS ?= -ldflags="-X main.Version=$(VERSION)"

.PHONY: all clean build-all build-anubis build-thoth build-maat build-scarab build-guard build-agent

all: build-all

# --- Standard Build ---
build-all: build-anubis build-thoth build-maat build-scarab build-guard build-agent

clean:
	rm -rf $(BUILD_DIR)

# --- Individual Deity Binaries ---
build-anubis:
	go build $(GO_FLAGS) -o $(BUILD_DIR)/anubis ./cmd/anubis/

build-thoth:
	go build $(GO_FLAGS) -o $(BUILD_DIR)/thoth ./cmd/thoth/

build-maat:
	go build $(GO_FLAGS) -o $(BUILD_DIR)/maat ./cmd/maat/

build-scarab:
	go build $(GO_FLAGS) -o $(BUILD_DIR)/scarab ./cmd/scarab/

build-guard:
	go build $(GO_FLAGS) -o $(BUILD_DIR)/guard ./cmd/guard/

build-agent:
	go build $(GO_FLAGS) -o $(BUILD_DIR)/pantheon-agent ./cmd/pantheon-agent/

# --- Public Proof of Testing ---
test-proof:
	go test -v -coverprofile=$(BUILD_DIR)/coverage.out ./...
	go tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "Public proof generated in $(BUILD_DIR)/coverage.html"
