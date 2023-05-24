## gerrit_related_changes
A CLI tool to find related changes given a Gerrit CL. See [Gerrit documentation](https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#get-related-changes).

### Sample usage

#### Related changes

`go run gerrit_related_changes.go gerrit_related_changes --input_json="sample-stacked.json" --output_json="/tmp/sample_gerrit_related_changes_output.txt"`

gives:

```
cat /tmp/sample_gerrit_related_changes_output.txt
{
 "related": [
  {
   "_change_number": 4508966
  },
  {
   "_change_number": 4508965
  }
 ],
 "relatedCount": 2,
 "hasRelated": true
```


#### Unrelated changes

`go run gerrit_related_changes.go gerrit_related_changes --input_json="sample-unstacked.json" --output_json="/tmp/sample_gerrit_related_changes_output.txt"`

gives:

```
cat /tmp/sample_gerrit_related_changes_output.txt
{
 "related": [],
 "relatedCount": 0,
 "hasRelated": false
}
```