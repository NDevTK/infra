Package private is for YOUR personal use. Do not check anything in besides this file.

Here is a sample function that might be useful but should NOT be checked in.

```
func Assert(cond bool, message string) {
	if !cond {
		panic(message)
	}
}
```
