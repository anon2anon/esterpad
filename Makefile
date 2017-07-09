all: utils deps backend frontend

frontend:
	$(MAKE) -C frontend
	
.PHONY: frontend

backend:
	GOPATH="$(CURDIR)" go build -o esterpad build.go
	mkdir -p log

deps:
	GOPATH="$(CURDIR)" go get -d

utils:
	protoc --go_out=. src/esterpad_utils/esterpad.proto

tester:
	GOPATH="$(CURDIR)" go build -o tester build_tester.go
