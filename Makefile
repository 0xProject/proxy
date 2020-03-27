.PHONY: proxy
proxy:
	go build && ./proxy

.PHONY: test
test: 
	go test ./...