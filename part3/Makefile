.PHONY: readtest optest asm bench test clean
all: optest
optest:
	@go run ./cmd/optest
readtest:
	@go run ./cmd/readtest ./tmp/output.json
asm:
	@nasm -f elf64 -o loops/loops.o loops/loops.asm
	@ar q loops/libloops.a loops/loops.o
	@rm loops/loops.o
bench:
	@go test -bench=. ./...
test:
	@go test -v ./...
clean:
	@rm -f ./generate ./haversine ./readtest
	@rm -f ./loops/libloops.a ./loops/loops.o
	@go clean
