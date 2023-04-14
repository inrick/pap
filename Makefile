.PHONY: all run test debug clean
all:
	go build
run:
	go run .
test:
	go test -v ./...
debug:
	gdlv debug
clean:
	go clean
