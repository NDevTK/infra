module infra

go 1.19

require (
	cloud.google.com/go v0.104.0
	cloud.google.com/go/appengine v1.2.0
	cloud.google.com/go/bigquery v1.42.0
	cloud.google.com/go/cloudtasks v1.6.0
	cloud.google.com/go/compute v1.10.0
	cloud.google.com/go/datastore v1.8.0
	cloud.google.com/go/firestore v1.6.1
	cloud.google.com/go/monitoring v1.5.0
	cloud.google.com/go/pubsub v1.25.1
	cloud.google.com/go/spanner v1.39.0
	cloud.google.com/go/storage v1.27.0
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v1.9.0
	github.com/PaesslerAG/jsonpath v0.1.1
	github.com/StackExchange/wmi v1.2.1
	github.com/VividCortex/godaemon v1.0.0
	github.com/aclements/go-moremath v0.0.0-20210112150236-f10218a38794
	github.com/alecthomas/participle/v2 v2.0.0-alpha7
	github.com/andygrunwald/go-gerrit v0.0.0-20210726065827-cc4e14e40b5b
	github.com/bazelbuild/remote-apis-sdks v0.0.0-20220518145948-4b34b65cbf91
	github.com/bmatcuk/doublestar v1.3.4
	github.com/containerd/cgroups v1.0.4
	github.com/danjacques/gofslock v0.0.0-20220131014315-6e321f4509c8
	github.com/docker/docker v20.10.8+incompatible
	github.com/dustin/go-humanize v1.0.0
	github.com/gofrs/flock v0.8.1
	github.com/gogo/protobuf v1.3.2
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2
	github.com/google/cel-go v0.7.3
	github.com/google/go-cmp v0.5.9
	github.com/google/go-containerregistry v0.6.0
	github.com/google/safetext v0.0.0-20220905092116-b49f7bc46da2
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/google/subcommands v1.2.0
	github.com/google/uuid v1.3.0
	github.com/googleapis/gax-go/v2 v2.5.1
	github.com/googleapis/google-cloud-go-testing v0.0.0-20210719221736-1c9a4c676720
	github.com/jdxcode/netrc v0.0.0-20210204082910-926c7f70242a
	github.com/julienschmidt/httprouter v1.3.0
	github.com/klauspost/compress v1.15.9
	github.com/kr/pretty v0.3.0
	github.com/kylelemons/godebug v1.1.0
	github.com/linkedin/goavro/v2 v2.11.0
	github.com/maruel/subcommands v1.1.1
	github.com/mitchellh/go-homedir v1.1.0
	github.com/opencontainers/image-spec v1.0.1
	github.com/otiai10/copy v1.6.0
	github.com/pkg/errors v0.9.1
	github.com/pkg/profile v1.6.0
	github.com/shirou/gopsutil v3.21.11+incompatible
	github.com/smartystreets/goconvey v1.7.2
	github.com/ulikunitz/xz v0.5.10
	github.com/waigani/diffparser v0.0.0-20190828052634-7391f219313d
	go.chromium.org/chromiumos/config/go v0.0.0-20211012171127-50826c369fec
	go.chromium.org/chromiumos/infra/proto/go v0.0.0-00010101000000-000000000000
	go.chromium.org/luci v0.0.0-20220907173800-9129d25df9a6
	go.opencensus.io v0.23.0
	go.opentelemetry.io/contrib/detectors/gcp v1.10.0
	go.opentelemetry.io/otel v1.10.0
	go.opentelemetry.io/otel/sdk v1.10.0
	go.opentelemetry.io/otel/trace v1.10.0
	go.skia.org/infra v0.0.0-20220916162150-5898edbc8e08
	golang.org/x/build v0.0.0-20210913192547-14e3e09d6b10
	golang.org/x/crypto v0.0.0-20220518034528-6f7dac969898
	golang.org/x/mobile v0.0.0-20191031020345-0945064e013a
	golang.org/x/net v0.0.0-20221012135044-0b7e1fb9d458
	golang.org/x/oauth2 v0.0.0-20221006150949-b44042a4b9c1
	golang.org/x/sync v0.0.0-20220929204114-8fcdb60fdcc0
	golang.org/x/sys v0.0.0-20220908150016-7ac13a9a928d
	golang.org/x/time v0.0.0-20220609170525-579cf78fd858
	golang.org/x/tools v0.1.12
	golang.org/x/tools/gopls v0.8.3
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2
	gonum.org/v1/gonum v0.9.3
	google.golang.org/api v0.99.0
	google.golang.org/appengine v1.6.8-0.20220805212354-d981f2f002d3
	google.golang.org/appengine/v2 v2.0.1
	google.golang.org/genproto v0.0.0-20221018160656-63c7b68cfc55
	google.golang.org/grpc v1.50.1
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.2.0
	google.golang.org/protobuf v1.28.1
	gopkg.in/fsnotify.v1 v1.4.7
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.4.0
	howett.net/plist v1.0.0
	k8s.io/api v0.22.12
	k8s.io/apimachinery v0.22.12
	k8s.io/client-go v0.22.12
	k8s.io/metrics v0.22.12
	sigs.k8s.io/yaml v1.3.0
)

require (
	cloud.google.com/go/errorreporting v0.2.0 // indirect
	cloud.google.com/go/iam v0.5.0 // indirect
	cloud.google.com/go/profiler v0.3.0 // indirect
	cloud.google.com/go/secretmanager v1.7.0 // indirect
	cloud.google.com/go/trace v1.2.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.13.12 // indirect
	github.com/BurntSushi/toml v1.0.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v0.33.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.33.0 // indirect
	github.com/Microsoft/go-winio v0.6.0 // indirect
	github.com/PaesslerAG/gval v1.0.0 // indirect
	github.com/antlr/antlr4/runtime/Go/antlr v0.0.0-20210907221601-4f80a5e09cd0 // indirect
	github.com/aws/aws-sdk-go v1.44.20 // indirect
	github.com/bazelbuild/remote-apis v0.0.0-20220510175640-3b4b64021035 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/census-instrumentation/opencensus-proto v0.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cncf/udpa/go v0.0.0-20220112060539-c52dc94e7fbe // indirect
	github.com/cncf/xds/go v0.0.0-20220520190051-1e77728a1eaa // indirect
	github.com/containerd/containerd v1.5.5 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/cli v20.10.7+incompatible // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.6.3 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/envoyproxy/go-control-plane v0.10.2-0.20220325020618-49ff273808a1 // indirect
	github.com/envoyproxy/protoc-gen-validate v0.6.13 // indirect
	github.com/fiorix/go-web v1.0.1-0.20150221144011-5b593f1e8966 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/godbus/dbus/v5 v5.0.4 // indirect
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gomodule/redigo v1.8.9 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/pprof v0.0.0-20220520215854-d04f2422c8a1 // indirect
	github.com/google/tink/go v1.7.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.0 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jcgregorio/logger v0.1.3 // indirect
	github.com/jcgregorio/slog v0.0.0-20190423190439-e6f2d537f900 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/luci/gtreap v0.0.0-20161228054646-35df89791e8f // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-tty v0.0.4 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mostynb/zstdpool-syncpool v0.0.12 // indirect
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/runtime-spec v1.0.3-0.20210326190908-1c3f411f0417 // indirect
	github.com/pborman/uuid v1.2.1 // indirect
	github.com/pkg/xattr v0.4.7 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.11.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.30.0 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/prometheus/prometheus v2.5.0+incompatible // indirect
	github.com/rogpeppe/go-internal v1.8.1 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/smartystreets/assertions v1.13.0 // indirect
	github.com/spf13/cobra v1.3.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/src-d/gcfg v1.4.0 // indirect
	github.com/stoewer/go-strcase v1.2.0 // indirect
	github.com/texttheater/golang-levenshtein v1.0.1 // indirect
	github.com/zeebo/bencode v1.0.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/exp/typeparams v0.0.0-20220218215828-6cf2b201936e // indirect
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4 // indirect
	golang.org/x/term v0.0.0-20220526004731-065cf7ba2467 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/vuln v0.0.0-20220324005316-18fd808f5c7f // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/src-d/go-billy.v4 v4.3.2 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	honnef.co/go/tools v0.3.0 // indirect
	k8s.io/klog/v2 v2.20.0 // indirect
	k8s.io/utils v0.0.0-20211116205334-6203023598ed // indirect
	mvdan.cc/gofumpt v0.3.0 // indirect
	mvdan.cc/xurls/v2 v2.4.0 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
)

// See https://github.com/google/cel-go/issues/441.
exclude github.com/antlr/antlr4 v0.0.0-20200503195918-621b933c7a7f

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
