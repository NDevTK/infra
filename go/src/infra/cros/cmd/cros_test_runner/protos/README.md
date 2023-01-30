## Protos for Cros Test Runner
This package should hold all the protos needed for cros test runner.

## Install protoc (if not already installed)
1. Check if protoc is installed in the env: `protoc --version`
2. If not installed, install protoc following these commands.
3. Run: `wget wget https://github.com/google/protobuf/releases/download/v3.3.0/protoc-3.3.0-linux-x86_64.zip`
4. Copy the binary to bin: `sudo cp bin/protoc  /usr/bin/protoc`
5. Run again: `protoc --version`. Output should be: `libprotoc 3.3.0`

## CipdInfo proto
Compile this proto via this command inside protos folder:
`protoc cipdInfo.proto --proto_path=./ --go_out=./`