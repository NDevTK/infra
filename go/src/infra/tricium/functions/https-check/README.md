# HTTPS-check for URLs

Tricium analyzer checking that all links (except `go/` and `g/` links)
in README files use `https`.

Consumes Tricium FILES and produces Tricium RESULTS comments.

## Development and Testing

Local testing:

```
$ go build
$ ./https-check --input=test --output=output
```
