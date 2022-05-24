# CPU Profiler for Chrome Build

This is CPU profiler periodically collecting CPU usage of local process during
Chrome build. This can be used to find build actions that may have low-hanging
fruit for build performance optimization.

## Install build_profiler from CIPD

```
$ echo 'infra/tools/build_profiler/${platform} latest' > ensure_file.txt
$ cipd ensure -ensure-file ensure_file.txt -root dir
```

## Run build_profiler

To run build\_profiler, you need to prepend build\_profiler for usual build
command.

e.g.

```
$ build_profiler autoninja -C out/Release chrome
```
