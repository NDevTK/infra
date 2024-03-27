module infra

go 1.22

require (
	cloud.google.com/go v0.112.0
	cloud.google.com/go/appengine v1.8.5
	cloud.google.com/go/bigquery v1.59.1
	cloud.google.com/go/cloudtasks v1.12.6
	cloud.google.com/go/compute v1.24.0
	cloud.google.com/go/compute/metadata v0.2.3
	cloud.google.com/go/datastore v1.15.0
	cloud.google.com/go/firestore v1.14.0
	cloud.google.com/go/logging v1.9.0
	cloud.google.com/go/longrunning v0.5.5
	cloud.google.com/go/monitoring v1.18.0
	cloud.google.com/go/profiler v0.4.0
	cloud.google.com/go/pubsub v1.36.1
	cloud.google.com/go/secretmanager v1.11.5
	cloud.google.com/go/storage v1.38.0
	cloud.google.com/go/trace v1.10.5
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/Microsoft/go-winio v0.6.1
	github.com/PaesslerAG/jsonpath v0.1.1
	github.com/StackExchange/wmi v1.2.1
	github.com/VividCortex/godaemon v1.0.0
	github.com/aclements/go-moremath v0.0.0-20210112150236-f10218a38794
	github.com/andygrunwald/go-gerrit v0.0.0-20210726065827-cc4e14e40b5b
	github.com/bazelbuild/reclient/api v0.0.0-20231027154936-0dffdbcf8db1
	github.com/bazelbuild/remote-apis v0.0.0-20240215191509-9ff14cecffe5
	github.com/bazelbuild/remote-apis-sdks v0.0.0-20240304155856-a108ad8bbbf4
	github.com/bmatcuk/doublestar v1.3.4
	github.com/containerd/cgroups v1.0.4
	github.com/danjacques/gofslock v0.0.0-20240212154529-d899e02bfe22
	github.com/docker/docker v24.0.7+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/dustin/go-humanize v1.0.1
	github.com/fsnotify/fsnotify v1.6.1-0.20230115155808-883002e0390d
	github.com/gliderlabs/ssh v0.3.5
	github.com/go-git/go-git/v5 v5.11.0
	github.com/gofrs/flock v0.8.1
	github.com/gogo/protobuf v1.3.2
	github.com/golang/glog v1.1.2
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.4
	github.com/google/cel-go v0.7.3
	github.com/google/go-cmp v0.6.0
	github.com/google/go-containerregistry v0.6.0
	github.com/google/safetext v0.0.0-20220905092116-b49f7bc46da2
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/google/subcommands v1.2.0
	github.com/google/uuid v1.6.0
	github.com/googleapis/gax-go/v2 v2.12.1
	github.com/googleapis/google-cloud-go-testing v0.0.0-20210719221736-1c9a4c676720
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-version v1.2.0
	github.com/jackc/pgx/v4 v4.18.2
	github.com/jdxcode/netrc v0.0.0-20210204082910-926c7f70242a
	github.com/julienschmidt/httprouter v1.3.0
	github.com/klauspost/compress v1.17.6
	github.com/kr/pretty v0.3.1
	github.com/kylelemons/godebug v1.1.0
	github.com/linkedin/goavro/v2 v2.11.0
	github.com/maruel/subcommands v1.1.1
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/hashstructure/v2 v2.0.2
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/opencontainers/image-spec v1.1.0-rc5
	github.com/opencontainers/runtime-spec v1.0.3-0.20210326190908-1c3f411f0417
	github.com/otiai10/copy v1.6.0
	github.com/pkg/errors v0.9.1
	github.com/pkg/profile v1.6.0
	github.com/pkg/xattr v0.4.9
	github.com/prometheus/client_golang v1.17.0
	github.com/savaki/jq v0.0.0-20161209013833-0e6baecebbf8
	github.com/shirou/gopsutil/v3 v3.22.12
	github.com/smartystreets/goconvey v1.8.1
	github.com/stretchr/testify v1.8.4
	github.com/ulikunitz/xz v0.5.10
	github.com/waigani/diffparser v0.0.0-20190828052634-7391f219313d
	go.chromium.org/chromiumos/config/go v0.0.0-20230102123018-be9aaedc3af8
	go.chromium.org/chromiumos/infra/proto/go v0.0.0-20240117203213-273a2bfa2e7e
	go.chromium.org/chromiumos/platform/dev-util/src/chromiumos/ctp v0.0.0-20240216220433-8a9cb838503d
	go.chromium.org/luci v0.0.0-20230103053340-8a57daa72e32
	go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace v0.47.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.47.0
	go.opentelemetry.io/otel v1.23.1
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.21.0
	go.opentelemetry.io/otel/sdk v1.23.1
	go.opentelemetry.io/otel/trace v1.23.1
	go.skia.org/infra v0.0.0-20230524042348-e1356ea7576d
	go.starlark.net v0.0.0-20240123142251-f86470692795
	golang.org/x/build v0.0.0-20210913192547-14e3e09d6b10
	golang.org/x/crypto v0.20.0
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d
	golang.org/x/mobile v0.0.0-20191031020345-0945064e013a
	golang.org/x/mod v0.15.0
	golang.org/x/net v0.21.0
	golang.org/x/oauth2 v0.17.0
	golang.org/x/perf v0.0.0-20210220033136-40a54f11e909
	golang.org/x/sync v0.6.0
	golang.org/x/sys v0.17.0
	golang.org/x/term v0.17.0
	golang.org/x/time v0.5.0
	golang.org/x/tools v0.18.0
	golang.org/x/xerrors v0.0.0-20231012003039-104605ab7028
	gonum.org/v1/gonum v0.12.0
	google.golang.org/api v0.164.0
	google.golang.org/appengine v1.6.8
	google.golang.org/appengine/v2 v2.0.4
	google.golang.org/genproto v0.0.0-20240213162025-012b6fc9bca9
	google.golang.org/genproto/googleapis/api v0.0.0-20240213162025-012b6fc9bca9
	google.golang.org/genproto/googleapis/bytestream v0.0.0-20240213162025-012b6fc9bca9
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240213162025-012b6fc9bca9
	google.golang.org/grpc v1.61.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.3.0
	google.golang.org/protobuf v1.33.0
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	gotest.tools v2.2.0+incompatible
	howett.net/plist v1.0.0
	k8s.io/api v0.22.12
	k8s.io/apimachinery v0.22.12
	k8s.io/client-go v0.22.12
	k8s.io/metrics v0.22.12
	sigs.k8s.io/yaml v1.3.0
)

require (
	cloud.google.com/go/errorreporting v0.3.0 // indirect
	cloud.google.com/go/iam v1.1.6 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.21.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v1.21.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.45.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/propagator v0.45.0 // indirect
	github.com/GoogleCloudPlatform/protoc-gen-bq-schema v0.0.0-20190119112626-026f9fcdf705 // indirect
	github.com/PaesslerAG/gval v1.0.0 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20230828082145-3c4c8a2d2371 // indirect
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect
	github.com/antlr/antlr4/runtime/Go/antlr v0.0.0-20210907221601-4f80a5e09cd0 // indirect
	github.com/apache/arrow/go/v14 v14.0.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cloudflare/circl v1.3.3 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/cyphar/filepath-securejoin v0.2.4 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/cli v20.10.7+incompatible // indirect
	github.com/docker/distribution v2.8.2+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.6.3 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.0.4 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fiorix/go-web v1.0.1-0.20150221144011-5b593f1e8966 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.5.0 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/godbus/dbus/v5 v5.0.6 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gomodule/redigo v1.8.9 // indirect
	github.com/google/flatbuffers v23.5.26+incompatible // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/pprof v0.0.0-20230602150820-91b7bce49751 // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/google/tink/go v1.7.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.16.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.14.3 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.3 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgtype v1.14.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jcgregorio/logger v0.1.3 // indirect
	github.com/jcgregorio/slog v0.0.0-20190423190439-e6f2d537f900 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/klauspost/cpuid/v2 v2.2.6 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/luci/gtreap v0.0.0-20161228054646-35df89791e8f // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mattn/go-tty v0.0.5 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mostynb/zstdpool-syncpool v0.0.12 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pborman/uuid v1.2.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.18 // indirect
	github.com/pjbgf/sha1cd v0.3.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/ahafs v0.1.4 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus/client_model v0.4.1-0.20230718164431-9a2bf3000d16 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.11.1 // indirect
	github.com/protocolbuffers/txtpbfmt v0.0.0-20240116145035-ef3ab179eed6 // indirect
	github.com/rogpeppe/go-internal v1.11.0 // indirect
	github.com/sergi/go-diff v1.3.1 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/smarty/assertions v1.15.1 // indirect
	github.com/spf13/cobra v1.3.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stoewer/go-strcase v1.3.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/texttheater/golang-levenshtein v1.0.1 // indirect
	github.com/tklauser/go-sysconf v0.3.11 // indirect
	github.com/tklauser/numcpus v0.6.0 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	github.com/zeebo/bencode v1.0.0 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.einride.tech/aip v0.66.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.23.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.48.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.21.0 // indirect
	go.opentelemetry.io/otel/metric v1.23.1 // indirect
	go.opentelemetry.io/proto/otlp v1.0.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	k8s.io/klog/v2 v2.20.0 // indirect
	k8s.io/utils v0.0.0-20211116205334-6203023598ed // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
)

// See https://github.com/google/cel-go/issues/441.
exclude github.com/antlr/antlr4 v0.0.0-20200503195918-621b933c7a7f

// Apparently checking out NDKs at head isn't really safe.
replace golang.org/x/mobile => golang.org/x/mobile v0.0.0-20170111200746-6f0c9f6df9bb

// Version 1.2.0 has a bug: https://github.com/sergi/go-diff/issues/115
exclude github.com/sergi/go-diff v1.2.0

// Infra modules are included via gclient DEPS.
replace (
	go.chromium.org/chromiumos/config/go => ../go.chromium.org/chromiumos/config/go/src/go.chromium.org/chromiumos/config/go
	go.chromium.org/chromiumos/infra/proto/go => ../go.chromium.org/chromiumos/infra/proto/go
	go.chromium.org/luci => ../go.chromium.org/luci
)
