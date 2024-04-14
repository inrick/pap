.PHONY: alignreadtest cachetestnonpow2 cachetest simdreadtest rwporttest optest readtest generate asm bench test clean
all: alignreadtest
alignreadtest: asm generate
	@go run ./cmd/alignreadtest
cachetestnonpow2: asm generate
	@go run ./cmd/cachetestnonpow2
cachetest: asm generate
	@go run ./cmd/cachetest
simdreadtest: asm generate
	@go run ./cmd/simdreadtest
rwporttest: asm
	@go run ./cmd/rwporttest
optest: asm
	@go run ./cmd/optest
readtest:
	@go run ./cmd/readtest ./tmp/output.json
generate:
	@go generate ./...
asm:
	@nasm -f elf64 -o asm/loops.o asm/loops.asm
	@ar q asm/libloops.a asm/loops.o
	@rm asm/loops.o
bench:
	@go test -bench=. ./...
test:
	@go test -v ./...
clean:
	@rm -f ./optest ./readtest
	@rm -f ./asm/libloops.a ./asm/loops.o
	@go clean
