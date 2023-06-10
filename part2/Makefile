.PHONY: all debuggen debughav test
all:
	@go run ./cmd/generate/generate.go -dir ./tmp cluster 1234 100
	@go run ./cmd/haversine/haversine.go ./tmp/output.json ./tmp/output.f64
debuggen:
	gdlv run ./cmd/generate/generate.go -dir ./tmp cluster 1234 1000
debughav:
	gdlv run ./cmd/haversine/haversine.go ./tmp/output.json ./tmp/output.f64
test:
	go test -v ./...
