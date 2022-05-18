module infra

go 1.17

require (
	cloud.google.com/go v0.100.2
	cloud.google.com/go/appengine v1.2.0
	cloud.google.com/go/bigquery v1.28.0
	cloud.google.com/go/cloudtasks v0.1.0
	cloud.google.com/go/compute v1.5.0
	cloud.google.com/go/datastore v1.5.0
	cloud.google.com/go/firestore v1.6.1
	cloud.google.com/go/monitoring v1.5.0
	cloud.google.com/go/pubsub v1.17.0
	cloud.google.com/go/spanner v1.29.0
	cloud.google.com/go/storage v1.21.0
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/PaesslerAG/jsonpath v0.1.1
	github.com/StackExchange/wmi v1.2.1
	github.com/VividCortex/godaemon v1.0.0
	github.com/aclements/go-moremath v0.0.0-20210112150236-f10218a38794
	github.com/alecthomas/participle/v2 v2.0.0-alpha7
	github.com/andygrunwald/go-gerrit v0.0.0-20210726065827-cc4e14e40b5b
	github.com/bazelbuild/remote-apis-sdks v0.0.0-20220422144733-8780f11b1bb2
	github.com/bmatcuk/doublestar v1.3.4
	github.com/danjacques/gofslock v0.0.0-20220131014315-6e321f4509c8
	github.com/docker/docker v20.10.8+incompatible
	github.com/dustin/go-humanize v1.0.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2
	github.com/google/cel-go v0.7.3
	github.com/google/go-cmp v0.5.7
	github.com/google/go-containerregistry v0.6.0
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/google/subcommands v1.2.0
	github.com/google/uuid v1.3.0
	github.com/googleapis/gax-go/v2 v2.3.0
	github.com/googleapis/google-cloud-go-testing v0.0.0-20210719221736-1c9a4c676720
	github.com/jdxcode/netrc v0.0.0-20210204082910-926c7f70242a
	github.com/julienschmidt/httprouter v1.3.0
	github.com/klauspost/compress v1.13.5
	github.com/kr/pretty v0.3.0
	github.com/kylelemons/godebug v1.1.0
	github.com/linkedin/goavro/v2 v2.11.0
	github.com/maruel/subcommands v1.1.1
	github.com/mattes/migrate v3.0.1+incompatible
	github.com/mitchellh/go-homedir v1.1.0
	github.com/opencontainers/image-spec v1.0.1
	github.com/otiai10/copy v1.6.0
	github.com/pkg/errors v0.9.1
	github.com/pkg/profile v1.6.0
	github.com/shirou/gopsutil v3.21.11+incompatible
	github.com/smartystreets/goconvey v1.7.2
	github.com/waigani/diffparser v0.0.0-20190828052634-7391f219313d
	go.chromium.org/chromiumos/config/go v0.0.0-20211012171127-50826c369fec
	go.chromium.org/chromiumos/infra/proto/go v0.0.0-00010101000000-000000000000
	go.chromium.org/luci v0.0.0-20201029184154-594d11850ebf
	go.opencensus.io v0.23.0
	go.skia.org/infra v0.0.0-20220517204524-3e1e25eccba0
	golang.org/x/build v0.0.0-20210913192547-14e3e09d6b10
	golang.org/x/crypto v0.0.0-20220112180741-5e0467b6c7ce
	golang.org/x/mobile v0.0.0-20191031020345-0945064e013a
	golang.org/x/net v0.0.0-20220325170049-de3da57026de
	golang.org/x/oauth2 v0.0.0-20220309155454-6242fa91716a
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20220328115105-d36c6a25d886
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac
	golang.org/x/tools v0.1.11-0.20220407163324-91bcfb1bdf9c
	golang.org/x/tools/gopls v0.8.3
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	gonum.org/v1/gonum v0.9.3
	google.golang.org/api v0.74.0
	google.golang.org/appengine v1.6.8-0.20220404170424-ca3835e9b4d6
	google.golang.org/appengine/v2 v2.0.1
	google.golang.org/genproto v0.0.0-20220426171045-31bebdecfb46
	google.golang.org/grpc v1.45.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0
	google.golang.org/protobuf v1.28.0
	gopkg.in/fsnotify.v1 v1.4.7
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.4.0
	howett.net/plist v0.0.0-20201203080718-1454fab16a06
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	k8s.io/metrics v0.22.1
	sigs.k8s.io/yaml v1.3.0
)

require (
	cloud.google.com/go/errorreporting v0.1.0 // indirect
	cloud.google.com/go/iam v0.3.0 // indirect
	cloud.google.com/go/profiler v0.1.0 // indirect
	cloud.google.com/go/secretmanager v1.4.0 // indirect
	cloud.google.com/go/trace v1.2.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.13.8 // indirect
	github.com/BurntSushi/toml v1.0.0 // indirect
	github.com/Microsoft/go-winio v0.5.0 // indirect
	github.com/PaesslerAG/gval v1.0.0 // indirect
	github.com/antlr/antlr4/runtime/Go/antlr v0.0.0-20210907221601-4f80a5e09cd0 // indirect
	github.com/aws/aws-sdk-go v1.40.42 // indirect
	github.com/bazelbuild/remote-apis v0.0.0-20210812183132-3e816456ee28 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/census-instrumentation/opencensus-proto v0.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cncf/udpa/go v0.0.0-20210930031921-04548b0d99d4 // indirect
	github.com/cncf/xds/go v0.0.0-20211130200136-a8f946100490 // indirect
	github.com/containerd/containerd v1.5.5 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/cli v20.10.7+incompatible // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.6.3 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/envoyproxy/go-control-plane v0.10.1 // indirect
	github.com/envoyproxy/protoc-gen-validate v0.6.7 // indirect
	github.com/fiorix/go-web v1.0.1-0.20150221144011-5b593f1e8966 // indirect
	github.com/go-logr/logr v1.1.0 // indirect
	github.com/go-ole/go-ole v1.2.5 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gomodule/redigo v1.8.5 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/pprof v0.0.0-20210827144239-02619b876842 // indirect
	github.com/google/tink/go v1.6.1 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20210901121439-eee08aaf2717 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jcgregorio/logger v0.1.2 // indirect
	github.com/jcgregorio/slog v0.0.0-20190423190439-e6f2d537f900 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/luci/gtreap v0.0.0-20161228054646-35df89791e8f // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-tty v0.0.3 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mostynb/zstdpool-syncpool v0.0.10 // indirect
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pborman/uuid v1.2.1 // indirect
	github.com/pkg/xattr v0.4.4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.11.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.30.0 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/protocolbuffers/txtpbfmt v0.0.0-20210910155415-af4447816e4d // indirect
	github.com/rogpeppe/go-internal v1.8.1 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/smartystreets/assertions v1.2.0 // indirect
	github.com/spf13/cobra v1.3.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/src-d/gcfg v1.4.0 // indirect
	github.com/stoewer/go-strcase v1.2.0 // indirect
	github.com/texttheater/golang-levenshtein v1.0.1 // indirect
	github.com/zeebo/bencode v1.0.0 // indirect
	golang.org/x/exp/typeparams v0.0.0-20220218215828-6cf2b201936e // indirect
	golang.org/x/mod v0.6.0-dev.0.20220106191415-9b9b3d81d5e3 // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/vuln v0.0.0-20220324005316-18fd808f5c7f // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/src-d/go-billy.v4 v4.3.2 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	honnef.co/go/tools v0.3.0 // indirect
	k8s.io/klog/v2 v2.20.0 // indirect
	k8s.io/utils v0.0.0-20210820185131-d34e5cb4466e // indirect
	mvdan.cc/gofumpt v0.3.0 // indirect
	mvdan.cc/xurls/v2 v2.4.0 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.1.2 // indirect
)

// See https://github.com/google/cel-go/issues/441.
exclude github.com/antlr/antlr4 v0.0.0-20200503195918-621b933c7a7f

// The following requirements were added on 2022-02-22 to prevent blocking deployment of
// go111 applications.
exclude google.golang.org/grpc v1.44.0

// The next version uses errors.Is(...) and no longer works on GAE go111.
replace golang.org/x/net => golang.org/x/net v0.0.0-20210503060351-7fd8e65b6420

// More recent versions break sysmon tests, crbug.com/1142700.
replace github.com/shirou/gopsutil => github.com/shirou/gopsutil v2.20.10-0.20201018091616-3202231bcdbd+incompatible

// Apparently checking out NDKs at head isn't really safe.
replace golang.org/x/mobile => golang.org/x/mobile v0.0.0-20170111200746-6f0c9f6df9bb

// Version 1.2.0 has a bug: https://github.com/sergi/go-diff/issues/115
exclude github.com/sergi/go-diff v1.2.0

// Infra modules are included via gclient DEPS.
replace (
	go.chromium.org/chromiumos/config/go => ../go.chromium.org/chromiumos/config/go
	go.chromium.org/chromiumos/infra/proto/go => ../go.chromium.org/chromiumos/infra/proto/go
	go.chromium.org/luci => ../go.chromium.org/luci
)
