.PHONY: all gen hav havrun debuggen debughav bench test clean
all: gen hav
	@#"https://github.com/golang/go/wiki/MinimumRequirements#amd64"
	@./generate -dir ./tmp cluster 1234 1000000
	@./haversine ./tmp/output.json ./tmp/output.f64
gen:
	@GOAMD64=v3 go build ./cmd/generate
hav:
	@go generate ./...
	@GOAMD64=v3 go build ./cmd/haversine
havrun: hav
	@time ./haversine ./tmp/output.json ./tmp/output.f64
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
