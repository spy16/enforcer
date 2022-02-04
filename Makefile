VERSION:=$(shell git describe --abbrev=0 --tags)
COMMIT:=$(shell git rev-list --abbrev-commit -1 HEAD)
BUILT_ON:=$(shell date +'%Y-%m-%d')

all: tidy test build

tidy:
	@echo "Tidying up..."
	@go mod tidy -v

test:
	@echo "Running unit tests..."
	@go test -cover ./...

build:
	@echo "Running go build..."
	@mkdir -p ./bin
	@go build -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuiltOn=$(BUILT_ON)" -o bin/ ./...
