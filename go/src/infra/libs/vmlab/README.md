# VMLab client library
Library for VMLab image syncing and instance leasing.

## Design doc
go/vmlab-client-leasing

## API definitions
The `api` package hosts all public APIs, including
- message protos: use `go generate ./api` to generate bindings in go
- public APIs
- prepopulated config data

## Code organization
- `vmlab.go` instantiates API implementations.
Inside internal:
- The `common` package hosts common utilities.
- The `image` package hosts implementations of image features: image sync.
- The `instance` package hosts implementation of instance features: VM leasing.

## Naming conventions
- VM, instance and VM instance are mostly interchangeable. Proto messages are
  named more verbosely as all messages are in the same package. Method names in
  interfaces are shortened as they are namespaced.
- All uppercase words are styled as regular words in the code. E.g. VmInstance
  instead of VMInstance.

## Usage example
```
import (
  "infra/libs/vmlab"
  "infra/libs/vmlab/api"
)
ins, err := vmlab.NewInstanceApi(api.ProviderId_GCLOUD)
if err != nil { ... }
ins.Create(&api.CreateVmInstanceRequest{ ... })
```
