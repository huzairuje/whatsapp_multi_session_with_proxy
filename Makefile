# WhatsApp Multi-Session Makefile
# Build and manage both backend (Go) and frontend (React)

# Variables
BACKEND_DIR = backend
FRONTEND_DIR = frontend
BIN_DIR = bin
BINARY_NAME = whatsapp_multi_session
BACKEND_BINARY = $(BIN_DIR)/$(BINARY_NAME)

# Go build flags
GO_BUILD_FLAGS = -ldflags="-w -s" -trimpath
CGO_ENABLED = 1

# Colors for output
COLOR_RESET = \033[0m
COLOR_BOLD = \033[1m
COLOR_GREEN = \033[32m
COLOR_YELLOW = \033[33m
COLOR_BLUE = \033[34m

.PHONY: help all clean install build-backend build-frontend build dev-backend dev-frontend dev test

# Default target
help:
	@echo "$(COLOR_BOLD)WhatsApp Multi-Session Build System$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_GREEN)Available targets:$(COLOR_RESET)"
	@echo "  $(COLOR_BOLD)make all$(COLOR_RESET)              - Install dependencies and build everything"
	@echo "  $(COLOR_BOLD)make install$(COLOR_RESET)          - Install all dependencies (Go + npm)"
	@echo "  $(COLOR_BOLD)make build$(COLOR_RESET)            - Build both backend and frontend"
	@echo "  $(COLOR_BOLD)make clean$(COLOR_RESET)            - Clean all build artifacts"
	@echo ""
	@echo "$(COLOR_YELLOW)Backend targets:$(COLOR_RESET)"
	@echo "  $(COLOR_BOLD)make build-backend$(COLOR_RESET)    - Build Go backend for Linux"
	@echo "  $(COLOR_BOLD)make build-windows$(COLOR_RESET)    - Build Go backend for Windows"
	@echo "  $(COLOR_BOLD)make dev-backend$(COLOR_RESET)      - Run backend in development mode"
	@echo "  $(COLOR_BOLD)make install-backend$(COLOR_RESET)  - Install Go dependencies"
	@echo ""
	@echo "$(COLOR_BLUE)Frontend targets:$(COLOR_RESET)"
	@echo "  $(COLOR_BOLD)make build-frontend$(COLOR_RESET)   - Build React frontend for production"
	@echo "  $(COLOR_BOLD)make dev-frontend$(COLOR_RESET)     - Run frontend in development mode"
	@echo "  $(COLOR_BOLD)make install-frontend$(COLOR_RESET) - Install npm dependencies"
	@echo ""

# Install all dependencies
install: install-backend install-frontend
	@echo "$(COLOR_GREEN)✓ All dependencies installed$(COLOR_RESET)"

install-backend:
	@echo "$(COLOR_YELLOW)Installing Go dependencies...$(COLOR_RESET)"
	cd $(BACKEND_DIR) && go mod download
	@echo "$(COLOR_GREEN)✓ Go dependencies installed$(COLOR_RESET)"

install-frontend:
	@echo "$(COLOR_YELLOW)Installing npm dependencies...$(COLOR_RESET)"
	cd $(FRONTEND_DIR) && npm install
	@echo "$(COLOR_GREEN)✓ npm dependencies installed$(COLOR_RESET)"

# Build everything
all: install build
	@echo "$(COLOR_GREEN)✓ Build complete!$(COLOR_RESET)"

build: build-backend build-frontend
	@echo "$(COLOR_GREEN)✓ Backend and frontend built successfully$(COLOR_RESET)"

# Backend build targets
build-backend: build-linux

build-linux:
	@echo "$(COLOR_YELLOW)Building backend for Linux (amd64)...$(COLOR_RESET)"
	@mkdir -p $(BIN_DIR)
	cd $(BACKEND_DIR) && env GOOS=linux GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) \
		go build $(GO_BUILD_FLAGS) -o ../$(BACKEND_BINARY)-linux-amd64
	@echo "$(COLOR_GREEN)✓ Linux binary: $(BACKEND_BINARY)-linux-amd64$(COLOR_RESET)"

build-windows:
	@echo "$(COLOR_YELLOW)Building backend for Windows (amd64)...$(COLOR_RESET)"
	@mkdir -p $(BIN_DIR)
	cd $(BACKEND_DIR) && env GOOS=windows GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) \
		CC=x86_64-w64-mingw32-gcc go build $(GO_BUILD_FLAGS) -o ../$(BACKEND_BINARY)-windows-amd64.exe
	@echo "$(COLOR_GREEN)✓ Windows binary: $(BACKEND_BINARY)-windows-amd64.exe$(COLOR_RESET)"

# Frontend build targets
build-frontend:
	@echo "$(COLOR_YELLOW)Building frontend for production...$(COLOR_RESET)"
	cd $(FRONTEND_DIR) && npm run build
	@echo "$(COLOR_GREEN)✓ Frontend built in $(FRONTEND_DIR)/dist$(COLOR_RESET)"

# Development targets
dev-backend: install-backend
	@echo "$(COLOR_YELLOW)Starting backend in development mode...$(COLOR_RESET)"
	cd $(BACKEND_DIR) && go run main.go

dev-frontend: install-frontend
	@echo "$(COLOR_YELLOW)Starting frontend in development mode...$(COLOR_RESET)"
	cd $(FRONTEND_DIR) && npm run dev

# Run both in development (requires terminal multiplexer or separate terminals)
dev:
	@echo "$(COLOR_YELLOW)To run both backend and frontend in development:$(COLOR_RESET)"
	@echo "  Terminal 1: make dev-backend"
	@echo "  Terminal 2: make dev-frontend"

# Clean targets
clean: clean-backend clean-frontend
	@echo "$(COLOR_GREEN)✓ All build artifacts cleaned$(COLOR_RESET)"

clean-backend:
	@echo "$(COLOR_YELLOW)Cleaning backend build artifacts...$(COLOR_RESET)"
	rm -rf $(BIN_DIR)
	cd $(BACKEND_DIR) && go clean
	@echo "$(COLOR_GREEN)✓ Backend cleaned$(COLOR_RESET)"

clean-frontend:
	@echo "$(COLOR_YELLOW)Cleaning frontend build artifacts...$(COLOR_RESET)"
	rm -rf $(FRONTEND_DIR)/dist
	rm -rf $(FRONTEND_DIR)/node_modules/.vite
	@echo "$(COLOR_GREEN)✓ Frontend cleaned$(COLOR_RESET)"

# Test targets (if tests exist)
test: test-backend test-frontend

test-backend:
	@echo "$(COLOR_YELLOW)Running backend tests...$(COLOR_RESET)"
	cd $(BACKEND_DIR) && go test ./... -v

test-frontend:
	@echo "$(COLOR_YELLOW)Running frontend tests...$(COLOR_RESET)"
	cd $(FRONTEND_DIR) && npm run test

# Setup config file
setup-config:
	@if [ ! -f $(BACKEND_DIR)/config.local.yaml ]; then \
		echo "$(COLOR_YELLOW)Creating config.local.yaml from example...$(COLOR_RESET)"; \
		cp config.local.yaml.example $(BACKEND_DIR)/config.local.yaml; \
		echo "$(COLOR_GREEN)✓ Config file created. Please edit $(BACKEND_DIR)/config.local.yaml$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_GREEN)✓ Config file already exists$(COLOR_RESET)"; \
	fi

# Quick start for new developers
quickstart: setup-config install build
	@echo ""
	@echo "$(COLOR_GREEN)$(COLOR_BOLD)✓ Quick start complete!$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_YELLOW)Next steps:$(COLOR_RESET)"
	@echo "  1. Edit $(BACKEND_DIR)/config.local.yaml with your settings"
	@echo "  2. Run backend:  make dev-backend"
	@echo "  3. Run frontend: make dev-frontend"
	@echo "  4. Open http://localhost:3000 in your browser"
	@echo ""
