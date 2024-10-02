test: 
	go test -failfast ./... -v

build:
	go build -o bin/ ./...

lint: ## Run golang-ci-lint to lint the go code (must `brew install golangci-lint` first)
	golangci-lint run