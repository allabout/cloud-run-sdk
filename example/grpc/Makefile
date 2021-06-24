.PHONY: go-proto
go-proto:
	protoc -I proto/ \
		--go_out=proto/ --go_opt=paths=source_relative \
		--go-grpc_out=proto/ --go-grpc_opt=paths=source_relative \
		proto/*.proto

.PHONY: run
run:
	go run ./server/main.go

.PHONY: run-client
run-client:
	go run ./client/main.go

.PHONY: run-client-fail
run-client-fail:
	go run ./client/main.go -message "ng word"