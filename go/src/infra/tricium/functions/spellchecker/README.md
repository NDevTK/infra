# Spellchecker

Tricium analyzer to check spelling on comments and text files.

Consumes Tricium FILES and produces Tricium RESULTS comments.

## Development and Testing

Local testing:

```
$ go build
$ ./spellchecker --input=test --output=out
```

## Deployment

Deploy a new version of the analyzer using CIPD:

```
$ make
$ cipd create -pkg-def=cipd.yaml
<outputs the VERSION>
$ cipd set-ref infra/tricium/function/spellchecker -ref live -version VERSION
```

## Adding Terms to Dictionary

The `dictionary.txt` file comes from the [`codespell`] repo. If you believe the
new terms you are adding is universally applicable, consider submit a PR to the
[`codespell`] repo and then sync the local copy using `make fetch-dict`.
Otherwise, you can add the new terms to `dictionary_extra.txt` to append or
override the terms in dictionary.

[`codespell`]: https://github.com/codespell-project/codespell
