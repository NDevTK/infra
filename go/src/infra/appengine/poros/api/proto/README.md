# Protocol Buffers for Poros

This folder contains the proto definitions about Poros data:

- resources.proto
- resource_service.proto
- asset_service.proto
- assetresource_service.proto
- asset_instance_service.proto

To generate Go code based on the proto definition, go to the appengine folder:

```shell
infra/go/src/infra/appengine$ protoc poros/api/proto/<xxx>.proto --go_out=../../ --go-grpc_out=../../ --proto_path=./
```

You may also use the opt option like the example in [grpc doc](https://grpc.io/docs/languages/go/quickstart/):

```shell
$ protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    helloworld/helloworld.proto
```
