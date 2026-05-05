.PHONY: install
install:
	go build -o $(GOPATH)/bin/rince ./cmd/rince/main.go
