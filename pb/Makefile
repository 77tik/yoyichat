all:gen

.PHONY: gen

gen:
	protoc connect.proto --go_out=./connect_pb --go_opt=paths=source_relative
	protoc logic.proto --go_out=./logic_pb --go_opt=paths=source_relative
	protoc task.proto --go_out=./task_pb --go_opt=paths=source_relative