.PHONY: build
build:
	go build -ldflags "-s -w" -o ./bin/gateway ./cmd

.PHONY: test
test:
	go test `go list ./...`

.PHONY: ci-testing
ci-testing:
	mkdir -p coverage
	go test -json -v -coverprofile ./coverage/coverage.txt `go list ./...`
	go tool cover -html=./coverage/coverage.txt -o ./coverage/coverage.html
