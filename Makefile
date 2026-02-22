.PHONY: build build-with-version run test clean

VERSION ?= v1.0.0
BUILD_DATE := $(shell date +'%Y/%m/%d %H:%M:%S')
BUILD_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

LDFLAGS := -X 'main.buildVersion=$(VERSION)' \
           -X 'main.buildDate=$(BUILD_DATE)' \
           -X 'main.buildCommit=$(BUILD_COMMIT)'

# Сборка без информации о версии
build:
	go build -o shortener.exe ./cmd/shortener

# Сборка с информацией о версии
build-with-version:
	@echo "Building with:"
	@echo "  Version: $(VERSION)"
	@echo "  Date: $(BUILD_DATE)"
	@echo "  Commit: $(BUILD_COMMIT)"
	@echo ""
	go build -ldflags "$(LDFLAGS)" -o shortener.exe ./cmd/shortener
	@echo "Build complete: shortener.exe"

# Запуск приложения
run:
	go run -ldflags "$(LDFLAGS)" ./cmd/shortener

# Запуск тестов
test:
	go test -v ./...

# Очистка
clean:
	rm -f shortener.exe reset.exe
	go clean -cache
