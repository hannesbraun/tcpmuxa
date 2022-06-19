sources := cmd/tcpmuxa/main.go config/*.go sysutils/*.go tcpmux/*.go

all: build

build: bin/tcpmuxa

bin/tcpmuxa: $(sources)
	go build -o bin/tcpmuxa cmd/tcpmuxa/main.go

run:
	go run cmd/tcpmuxa/main.go

clean:
	rm -r bin

.PHONY: all build clean run
