APP_NAME=Go-Blog-APP
BIN_NAME=go-bin
BUILD_DIR=./bin
GO_FILES=$(shell find . -name '*.go' -not -path '*/vendor/*')

ENV?=development
PORT?=8080
GO_VER=$(shell go version | awk '{print $$3}')

# ANSI Colors
BOLD=\033[1m
CYAN=\033[36m
YELLOW=\033[33m
GREEN=\033[32m
RESET=\033[0m

build:
	@echo "$(CYAN)🔨 BUILD: Building binary...$(RESET)"
	@go build -o $(BUILD_DIR)/$(BIN_NAME)
	@echo "$(RESET)"
	@echo ""
	@echo "✅ Binary built successfully!"
	@echo "Path: $(BUILD_DIR)/$(BIN_NAME)"
	@echo ""

clean:
	@echo "$(CYAN)🧹 CLEAN: Cleaning up...$(RESET)"
	@rm -rf $(BUILD_DIR)
	@echo "$(RESET)"
	@echo ""
	@echo "✅ Cleanup complete!"
	@echo ""

deps:
	@echo "$(CYAN)📥 DEPS: Installing dependencies...$(RESET)"
	@go mod tidy
	@echo "$(RESET)"
	@echo ""

fmt:
	@echo "$(CYAN)🧹🧹 FMT: Formatting code...$(RESET)"
	@go fmt ./...
	@echo "$(RESET)"
	@echo ""

help:
	@echo "$(CYAN)📜 HELP: Available commands...$(RESET)"
	@echo ""
	@echo "  make build      - Build the binary"
	@echo "  make clean      - Clean up build artifacts"
	@echo "  make deps       - Install dependencies"
	@echo "  make fmt        - Format code"
	@echo "  make help       - Show this help message"
	@echo "  make migrate-up - Run database migrations"
	@echo "  make migrate-down - Revert/rollback database migrations"
	@echo "  make run        - Run the application"
	@echo ""
	@echo "✅ Help command executed!"
	@echo ""

migrate-up:
	@echo "$(CYAN)🔄 MIGRATE: Running database migrations...$(RESET)"
	@dbmate up
	@echo "$(RESET)"
	@echo ""

migrate-down:
	@echo "$(CYAN)🔄 MIGRATE: Reverting database migrations...$(RESET)"
	@dbmate down
	@echo "$(RESET)"
	@echo ""

run:
	@sleep 1
	@echo "$(CYAN)⚙️ PREPARE: Checking system specs...$(RESET)"
	@sleep 1
	@echo "▶ Go Version : $(BOLD)$(GO_VER)$(RESET)"
	@echo "▶ Environment: $(BOLD)$(ENV)$(RESET)"
	@echo "▶ Server URL : $(BOLD)http://localhost:$(PORT)$(RESET)"
	@sleep 1
	@echo "$(YELLOW)⏳ GET READY: Countdown sequence initiated ..."
	@echo "3"
	@sleep 1
	@echo "2"
	@sleep 1
	@echo "1"
	@sleep 1
	@echo "$(RESET)"
	@echo ""
	@echo "$(GREEN)🚀 T-MINUS ZERO! WE HAVE LIFT OFF! TAKE OFF!$(RESET)"
	@echo ""
	@echo "Launching $(APP_NAME) server...$(RESET)"
	@echo ""
	@docker compose up --build
