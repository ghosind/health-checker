.PHONY: all

all:
	env GOOS=linux GOARCH=arm go build
