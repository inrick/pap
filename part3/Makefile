.PHONY: readtest optest asm bench test clean
all: optest
optest: asm
	@go run ./cmd/optest
readtest:
	@go run ./cmd/readtest ./tmp/output.json
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
