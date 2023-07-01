.PHONY: all gen hav debuggen debughav bench test clean
all:
	@#"https://github.com/golang/go/wiki/MinimumRequirements#amd64"
	@GOAMD64=v3 go build ./cmd/generate
	@./generate -dir ./tmp cluster 1234 100000
	@GOAMD64=v3 go build ./cmd/haversine
	@./haversine ./tmp/output.json ./tmp/output.f64
gen:
	@GOAMD64=v3 go build ./cmd/generate
hav:
	@GOAMD64=v3 go build ./cmd/haversine
debuggen:
	@go build ./cmd/generate
	@gdlv exec ./generate -dir ./tmp cluster 1234 1000
debughav:
	@go build ./cmd/haversine
	@gdlv exec ./haversine ./tmp/output.json ./tmp/output.f64
bench:
	@go test -bench=. ./...
test:
	@go test -v ./...
clean:
	@rm -f ./generate ./haversine
	@go clean
