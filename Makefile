.PHONY: all

all:
	env GOOS=linux GOARCH=amd64 go build
