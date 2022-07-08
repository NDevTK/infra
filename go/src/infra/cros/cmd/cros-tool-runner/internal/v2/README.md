# cros-tool-runner v2 Container Services

Container services expose certain docker commands as gRPC services.

## Design doc
go/cft-container-services

## API definition
For now, API protos are defined in `api` package under `cros-tool-runner`.
Use `go generate` to generate code for go.
`cros-tool-runner$ go generate ./api`

## Code organization
- `commands` package adopts command design pattern to provide an abstraction
  layer to interact with CLI tools.
- `server` package hosts the service implementation.
- TBD: `extensions` package implements extensions to the generic services.
  Namely, easy-to-use "templated" APIs for commonly used containers.
- Within a package, utility methods are grouped under a struct to be
  differentiated from other more "important" methods.

## Run locally
```shell
cros-tool-runner$ go build .
cros-tool-runner$ ./cros-tool-runner serve
server listening at [::]:8082
```
Once the server has started, you may use `grpc_cli` to interact with the
services. Example:
```shell
$ grpc_cli ls localhost:8082 ctrv2.api.CrosToolRunnerContainerService --channel_creds_type insecure
CreateNetwork
GetNetwork
Shutdown
StartContainer
$ grpc_cli call localhost:8082 ctrv2.api.CrosToolRunnerContainerService.GetNetwork "name: 'bridge'" --channel_creds_type insecure
connecting to localhost:8082
network {
  name: "bridge"
  id: "55be43ae262d69f75750675c9f6491a11271e8afbcc80d3c21a162d8b7d24831"
}
Rpc succeeded with OK status
$ grpc_cli call localhost:8082 ctrv2.api.CrosToolRunnerContainerService.StartContainer "name: 'my_container' container_image: 'us-docker.pkg.dev/cros-registry/test-services/cros-dut:8811903382633993457' start_command: ['cros-dut', '-cache_address', 'localhost:7443', '-dut_address', 'dut_address:2222']" --channel_creds_type insecure
connecting to localhost:8082
container {
  name: "my_container"
  id: "3d67e5d71e55f16e7535d0a21f766e9daf919b280232dd3d05faddeb574aaa27"
  owned: true
}
Rpc succeeded with OK status
```
