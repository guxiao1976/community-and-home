.PHONY: build test run-rpc run-api clean

build:
	go build ./...

test:
	go test ./...

run-rpc:
	cd rpc && go run communityhub.go -f etc/communityhub.yaml

run-api:
	cd api && go run communityhub.go -f etc/communityhub-api.yaml

clean:
	go clean ./...
