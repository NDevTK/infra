# SatLab

### Requirement
- .service_account.json: The credential is used to validate the GCS Bucket connection, and we can get this credential by logging in Google's partner account.
- SSH rsa_key: The credential is used to validate the SSH connection, and we need to set the `keys` path before building the `SatLab Server`.

### How to config `SSH rsa key` path
- `SSHKeyPath` needs to be on the path set in `utils/constants/constants.go`

### Build
Before running `go build`, we need to confirm the environment.

- .service_account.json
- SSH RSA key
