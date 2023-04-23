.PHONY: all run test debug clean
all:
	go build
run:
	go run . -exec -assemble
test:
	go test -v ./...
debug:
	gdlv debug -exec -assemble
clean:
	go clean
