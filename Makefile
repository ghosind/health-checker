.PHONY:

linux_x64:
	env GOOS=linux GOARCH=amd64 go build
