# cros-tool-runner v2 Container Services

Container services expose certain docker commands as gRPC services.

## Design doc
go/cft-container-services

## API definition
API protos are located in `src/config/proto/chromiumos/test/api` along with the
rest test API definitions.

## Code organization
- `commands` package adopts command design pattern to provide an abstraction
  layer to interact with CLI tools.
- `server` package hosts the service implementation.
- `templates` package implements extensions to the generic services.
  Namely, easy-to-use "templated" APIs for commonly used containers.
- Within a package, utility methods are grouped under a struct to be
  namespaced from other more "important" methods.

## Run locally
```shell
cros-tool-runner$ go build .
cros-tool-runner$ ./cros-tool-runner serve
server listening at [::]:8082
```
Once the server has started, you may use [`grpc_cli`](http://go/grpc_cli) to
interact with the services. Example:
```shell
$ grpc_cli ls localhost:8082 chromiumos.test.api.CrosToolRunnerContainerService --channel_creds_type insecure
CreateNetwork
GetNetwork
Shutdown
LoginRegistry
StartContainer
StartTemplatedContainer
StackCommands
$ grpc_cli call localhost:8082 chromiumos.test.api.CrosToolRunnerContainerService.GetNetwork "name: 'bridge'" --channel_creds_type insecure
connecting to localhost:8082
network {
  name: "bridge"
  id: "55be43ae262d69f75750675c9f6491a11271e8afbcc80d3c21a162d8b7d24831"
}
Rpc succeeded with OK status
$ grpc_cli call localhost:8082 chromiumos.test.api.CrosToolRunnerContainerService.LoginRegistry "username: 'oauth2accesstoken' password: '$(gcloud auth print-access-token)' registry: 'us-docker.pkg.dev'" --channel_creds_type insecure
connecting to localhost:8082
message: "Login Succeeded"
Rpc succeeded with OK status
$ grpc_cli call localhost:8082 chromiumos.test.api.CrosToolRunnerContainerService.StartContainer "name: 'my_container' container_image: 'us-docker.pkg.dev/cros-registry/test-services/cros-dut:8811903382633993457' start_command: ['cros-dut', '-cache_address', 'localhost:7443', '-dut_address', 'dut_address:2222']" --channel_creds_type insecure
connecting to localhost:8082
container {
  name: "my_container"
  id: "3d67e5d71e55f16e7535d0a21f766e9daf919b280232dd3d05faddeb574aaa27"
  owned: true
}
Rpc succeeded with OK status
```
* Depending on your setup, you may need to run `gcloud auth login` to be able to
  call the LoginRegistry endpoint.
* You may use the `--infile` flag of grpc_cli to read request data from a
  textproto file. See data/v2_example.textproto
