# Export CGO_ENABLED=1
CGO_ENABLED=1
# Export C compiler
CC=gcc

build:
	go build -ldflags "-H windowsgui" -o bin/