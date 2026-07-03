test:
	go test ./...

build:
	go build ./cmd/ch

build-windows:
	mkdir -p dist
	GOOS=windows GOARCH=amd64 go build -o dist/ch.exe ./cmd/ch

fmt:
	gofmt -w cmd/ch/main.go internal/cli/*.go internal/config/*.go internal/logging/*.go internal/safety/*.go internal/version/*.go
