go get -u github.com/golang/protobuf/protoc-gen-go 
protoc -I network network/network.proto --go_out=paths=source_relative:network
