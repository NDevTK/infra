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

## Login
```
./mallet login
```

## Example use of mallet: Running auto-repair from local code
```
go run main.go local -dev-jump-host [ip address of dut]
```

## Scheduling a job with mallet

This command also runs auto-repair, but instead of running locally,
it schedules a job to the DUT and generates both a luci job and a
swarming task page through labpack_builder.

```
go run main.go config > generated_config.json
./mallet recovery -config generated_config.json [DUT_NAME]
```

[Go env setup]: http://go/fleet-creating-work-env
