.PHONY: all run test debug clean
all:
	go build
run:
	go run . -exec
test:
	go test -v ./...
debug:
	gdlv debug -exec
clean:
	go clean
