run:
	GOPATH="$(shell pwd)" go run main.go

ngrok:
	ngrok start -config ngrok.yaml ng
