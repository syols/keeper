build:
	echo "Build agent"
	go build -o bin/client cmd/client/main.go
	echo "Build agent"
	go build -o bin/server cmd/server/main.go

server:
	go run cmd/server/main.go

client:
	go run cmd/client/main.go

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

run: server
