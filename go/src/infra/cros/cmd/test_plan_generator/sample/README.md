# Test plan generator sample

So you want to try running the test plan generator locally? Cool! Alright,
you're going to have to do a bit of setup, and you'll have to be a Googler.

1. Have the infra/infra repo (you're in it now) checked out (see
   https://chromium.googlesource.com/infra/infra/).
1. Set up the environment for Golang development in the infra/infra repo using the instructions [here](https://chromium.googlesource.com/infra/infra/+/refs/heads/main/go/README.md) and navigate to the `go/src/infra/cros` directory.
1. Have depot_tools (in particular, the repo command) on your PATH.

OK, now you can edit gen_test_plan_input.json if desired. Because of the serialized proto fields, the only way to change the contents presently is by grabbing the input from an actual build ([example](https://screenshot.googleplex.com/9ctgDtZmtdzWa9s)).

***
TODO(b/234019083): Create a script/workflow to create a sample input JSON file. The `serialized_proto` fields aren't editable.
<!-- Can convert with commands like `gqui from textproto:build_bucket_build_1.textproto proto buildbucket.v2.Build --outfile=rawproto:build_bucket_build_1.binarypb` -->
***

This might look something like

```json
{
  "gitiles_commit": "SomeSerializedGitilesCommit",
  "buildbucket_protos": [
    {
      "serialized_proto": "SomeSerializedBuildBucketBuildProto"
    },
    {
      "serialized_proto": "SomeOtherSerializedBuildBucketBuildProto"
    }
  ]
}
```

Alright, now you'll need to get OAuth credentials to run the program:

```shell
# Run from the go/src/infra/cros directory:
go run ./cmd/test_plan_generator/ auth-login
```

And now you can actually run it:

```shell
go run ./cmd/test_plan_generator/ gen-test-plan \
    --input_json=$PWD/cmd/test_plan_generator/sample/gen_test_plan_input.json \
    --output_json=$PWD/cmd/test_plan_generator/sample/gen_test_plan_output.json
```
