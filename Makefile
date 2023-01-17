build:
	echo "Build agent"
	go build -o bin/agent cmd/agent/main.go
	echo "Build agent"
	go build -o bin/server cmd/server/main.go

server:
	go run cmd/server/main.go

agent:
	go run cmd/agent/main.go

imports:
	goimports -l -w .

fmt:
	go fmt ./...

lint:
	golint ./...

vet:
	go vet -v ./...

errors:
	errcheck -ignoretests -blank ./...

deps:
	godep restore

test: deps
	go test -v ./...

coverage:
	go test -v -coverpkg=./... -coverprofile=profile.cov ./...
	go tool cover -func profile.cov

run: server
