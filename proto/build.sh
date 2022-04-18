protoc -I . --plugin="protoc-gen-go=${GOPATH}/bin/protoc-gen-go.exe" --plugin="protoc-gen-micro=${GOPATH}/bin/protoc-gen-micro.exe" --go_out=. --micro_out=. --go_opt=paths=source_relative --proto_path=. proto.proto


