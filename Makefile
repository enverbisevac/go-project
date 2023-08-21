ifndef GOBIN
	GOBIN := $(shell go env GOPATH)/bin
endif

BINARY_NAME=sample

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

.PHONY: all
all: help

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help: ## Show this help.
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)

# ==================================================================================== #
# Tests
# ==================================================================================== #

e2e: install
	@sh prep-e2e.sh
	@cd tests && venom run --var-from-file=./vars.yaml && rm *.log
	@pkill app

e2e-dev:
	@cd tests && venom run --var-from-file=./vars-dev.yaml && rm *.log
	@pkill app

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## tidy: format code and tidy modfile
.PHONY: tidy
tidy:
	go mod tidy -v

# Format go code and error if any changes are made
.PHONY: format
format: $(GOBIN)/goimports ## Format files using goimports
	@echo "Runing gofumpt"
	@gofumpt -l -w .
	@echo "Running goimports"
	@test -z $$(goimports -w ./..) || (echo "goimports would make a change. Please verify and commit the proposed changes"; exit 1)

.PHONY: lint
lint: $(GOBIN)/golangci-lint ## Run golangci-lint
	@golangci-lint run -v

## audit: run quality control checks
.PHONY: audit
audit:
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go test -race -vet=off ./...
	go mod verify


# ==================================================================================== #
# BUILD
# ==================================================================================== #

## build: build the cmd/api application
.PHONY: build
build: clean
	go mod verify
	CGO_ENABLED=1 go build -ldflags='-s' -o=./bin/${BINARY_NAME} ./cmd/app

clean:
	@echo "Remove binaries"
	@go clean
	@rm -rf ./bin

## init: initialize with sample data
.PHONY: init
init: tidy install
	@app init -t=restaurant


## run: run the cmd/api application
.PHONY: run
run: tidy build
	./bin/${BINARY_NAME} server

## install: installs sample application
.PHONY: install
install:
	@echo "Installing app"
	@go install ./cmd/app

# ==================================================================================== #
# TOOLS
# ==================================================================================== #
$(GOBIN)/golangci-lint:
	@echo "🔘 Installing golangci-lint... (`date '+%H:%M:%S'`)"
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN)

$(GOBIN)/goimports:
	@echo "🔘 Installing goimports ... (`date '+%H:%M:%S'`)"
	@go install golang.org/x/tools/cmd/goimports@latest

$(GOBIN)/gofumpt:
	@echo "🔘 Installing gofumpt ... (`date '+%H:%M:%S'`)"
	@go install mvdan.cc/gofumpt@latest

$(GOBIN)/gci:
	@echo "🔘 Installing gci ... (`date '+%H:%M:%S'`)"
	@go install github.com/daixiang0/gci@latest

tools = $(addprefix $(GOBIN)/, golangci-lint goimports gofumpt gci)

.PHONY: tools
tools: $(tools) ## Install tools

.PHONY: update-tools
update-tools: delete-tools tools ## Update tools

.PHONY: delete-tools
delete-tools: ## Delete the tools
	@rm $(tools) || true