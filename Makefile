all: deps backend frontend

frontend:
	$(MAKE) -C frontend
	
.PHONY: frontend

backend:
	protoc --go_out=. src/esterpad/clientmessages.proto
	GOPATH="$(CURDIR)" go build -o esterpad build.go
	mkdir -p log

deps:
	GOPATH="$(CURDIR)" go get -d
