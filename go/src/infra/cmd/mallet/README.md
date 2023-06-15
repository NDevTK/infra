# Mallet command line tool

This is a tool for running ad hoc maintenance tasks
that bypasses the ordinary infrastructure.

## Requirements
You should have Go installed. Refer to [Go env setup] to set up your
environment.

## Make and build
Run the following commands:
```
go mod tidy
go build
```

## Example use of mallet: Running auto-repair from local code
```
go run main.go local -dev-jump-host [ip address of dut]
```

[Go env setup]: http://go/fleet-creating-work-env
