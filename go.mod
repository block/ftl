module github.com/block/ftl

go 1.24.0

require (
	connectrpc.com/connect v1.17.0
	connectrpc.com/grpcreflect v1.3.0
	connectrpc.com/otelconnect v0.7.2
	github.com/BurntSushi/toml v1.5.0
	github.com/IBM/sarama v1.45.1
	github.com/PaesslerAG/gval v1.0.0
	github.com/XSAM/otelsql v0.38.0
	github.com/alecthomas/assert/v2 v2.11.0
	github.com/alecthomas/atomic v0.1.0-alpha2
	github.com/alecthomas/chroma/v2 v2.18.0
	github.com/alecthomas/concurrency v0.0.2
	github.com/alecthomas/errors v0.8.2
	github.com/alecthomas/kong v1.11.0
	github.com/alecthomas/kong-toml v0.2.0
	github.com/alecthomas/participle/v2 v2.1.4
	github.com/alecthomas/types v0.20.0
	github.com/amacneil/dbmate/v2 v2.27.0
	github.com/aws/aws-sdk-go-v2 v1.36.3
	github.com/aws/aws-sdk-go-v2/config v1.29.14
	github.com/aws/aws-sdk-go-v2/credentials v1.17.67
	github.com/aws/aws-sdk-go-v2/service/ecr v1.44.0
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.35.4
	github.com/aws/smithy-go v1.22.3
	github.com/beevik/etree v1.5.1
	github.com/block/ftl-golang-tools v0.5.0
	github.com/block/ftl-mysql-auth-proxy v0.0.0-20250514012428-f17e23827cde
	github.com/block/scaffolder v1.5.1
	github.com/bmatcuk/doublestar/v4 v4.8.1
	github.com/deckarep/golang-set/v2 v2.8.0
	github.com/docker/docker v28.1.1+incompatible
	github.com/docker/go-connections v0.5.0
	github.com/go-logr/logr v1.4.2
	github.com/go-openapi/spec v0.21.0
	github.com/go-sql-driver/mysql v1.9.2
	github.com/go-viper/mapstructure/v2 v2.2.1
	github.com/google/go-cmp v0.7.0
	github.com/google/go-containerregistry v0.20.5
	github.com/google/uuid v1.6.0
	github.com/hashicorp/cronexpr v1.1.2
	github.com/jackc/pgx/v5 v5.7.5
	github.com/jotaen/kong-completion v0.0.6
	github.com/jpillora/backoff v1.0.0
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/lni/dragonboat/v4 v4.0.0-20240618143154-6a1623140f27
	github.com/mark3labs/mcp-go v0.30.0
	github.com/mattn/go-isatty v0.0.20
	github.com/multiformats/go-base36 v0.2.0
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.1.1
	github.com/otiai10/copy v1.14.1
	github.com/posener/complete v1.2.3
	github.com/radovskyb/watcher v1.0.7
	github.com/rs/cors v1.11.1
	github.com/santhosh-tekuri/jsonschema/v5 v5.3.1
	github.com/sergi/go-diff v1.3.1
	github.com/swaggest/jsonschema-go v0.3.78
	github.com/tidwall/pretty v1.2.1
	github.com/titanous/json5 v1.0.0
	github.com/tliron/commonlog v0.2.19
	github.com/tliron/glsp v0.2.2
	github.com/tliron/kutil v0.3.26
	github.com/zalando/go-keyring v0.2.6
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.61.0
	go.opentelemetry.io/otel v1.36.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.36.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.36.0
	go.opentelemetry.io/otel/metric v1.36.0
	go.opentelemetry.io/otel/sdk v1.36.0
	go.opentelemetry.io/otel/sdk/metric v1.36.0
	go.opentelemetry.io/otel/trace v1.36.0
	go.uber.org/automaxprocs v1.6.0
	golang.org/x/exp v0.0.0-20250408133849-7e4ce0ab07d0
	golang.org/x/mod v0.24.0
	golang.org/x/net v0.40.0
	golang.org/x/sync v0.14.0
	golang.org/x/term v0.32.0
	golang.org/x/tools v0.33.0
	google.golang.org/protobuf v1.36.6
	gotest.tools/v3 v3.5.2
	istio.io/api v1.26.0
	istio.io/client-go v1.26.0
	k8s.io/api v0.33.1
	k8s.io/apimachinery v0.33.1
	k8s.io/client-go v0.33.1
	modernc.org/sqlite v1.37.1
	oras.land/oras-go/v2 v2.6.0
	sigs.k8s.io/yaml v1.4.0
)

require (
	al.essio.dev/pkg/shellescape v1.5.1 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/DataDog/zstd v1.5.2 // indirect
	github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/VictoriaMetrics/metrics v1.18.1 // indirect
	github.com/armon/go-metrics v0.0.0-20180917152333-f0300d1749da // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.30 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.25.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.30.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.19 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/cenkalti/backoff/v5 v5.0.2 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cockroachdb/errors v1.9.1 // indirect
	github.com/cockroachdb/logtags v0.0.0-20211118104740-dabe8e521a4f // indirect
	github.com/cockroachdb/pebble v0.0.0-20221207173255-0f086d933dac // indirect
	github.com/cockroachdb/redact v1.1.3 // indirect
	github.com/containerd/stargz-snapshotter/estargz v0.16.3 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/cli v28.1.1+incompatible // indirect
	github.com/docker/distribution v2.8.3+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.9.3 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/eapache/go-resiliency v1.7.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230731223053-c322873962e3 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/getsentry/sentry-go v0.12.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/gnostic-models v0.6.9 // indirect
	github.com/google/go-github/v71 v71.0.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.0.0 // indirect
	github.com/hashicorp/go-msgpack/v2 v2.1.1 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-sockaddr v1.0.0 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/hashicorp/memberlist v0.5.1 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lni/goutils v1.4.0 // indirect
	github.com/lni/vfs v0.2.1-0.20220616104132-8852fd867376 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/miekg/dns v1.1.26 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/sys/sequential v0.6.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/muesli/termenv v0.15.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/otiai10/mint v1.6.3 // indirect
	github.com/petermattis/goid v0.0.0-20240813172612-4fcff4a6cae7 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pingcap/errors v0.11.5-0.20240311024730-e056997136bb // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/riywo/loginshell v0.0.0-20200815045211-7d26008be1ab // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	github.com/sasha-s/go-deadlock v0.3.5 // indirect
	github.com/sean-/seed v0.0.0-20170313163322-e2103e2c3529 // indirect
	github.com/segmentio/ksuid v1.0.4 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/sourcegraph/jsonrpc2 v0.2.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/valyala/fastrand v1.1.0 // indirect
	github.com/valyala/histogram v1.2.0 // indirect
	github.com/vbatts/tar-split v0.12.1 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/proto/otlp v1.6.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/time v0.11.0 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20250318190949-c8a335a9a2ff // indirect
	k8s.io/utils v0.0.0-20241104100929-3ea5e8cea738 // indirect
	modernc.org/libc v1.65.7 // indirect
	sigs.k8s.io/json v0.0.0-20241010143419-9aa6b5e7a4b3 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.6.0 // indirect
)

require (
	github.com/PaesslerAG/jsonpath v0.1.1
	github.com/akamensky/base58 v0.0.0-20210829145138-ce8bf8802e8f
	github.com/alecthomas/repr v0.4.0
	github.com/aws/aws-sdk-go-v2/feature/rds/auth v1.5.11
	github.com/aws/aws-sdk-go-v2/service/cloudformation v1.59.2
	github.com/awslabs/goformation/v7 v7.14.9
	github.com/benbjohnson/clock v1.3.5
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/chzyer/readline v1.5.1
	github.com/danieljoos/wincred v1.2.2 // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/dop251/goja v0.0.0-20250309171923-bcd7cc6bf64c // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-sourcemap/sourcemap v2.1.3+incompatible // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/google/go-github/v72 v72.0.0
	github.com/google/pprof v0.0.0-20250317173921-a4b03ec1a45e // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3 // indirect
	github.com/hexops/gotextdiff v1.0.3
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/migueleliasweb/go-github-mock v1.3.0
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/puzpuzpuz/xsync/v3 v3.5.1
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/swaggest/refl v1.4.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.36.0 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/sys v0.33.0
	golang.org/x/text v0.25.0
	google.golang.org/genproto/googleapis/api v0.0.0-20250519155744-55703ea1f237 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250519155744-55703ea1f237 // indirect
	google.golang.org/grpc v1.72.1 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
)

retract (
	v1.1.5
	v1.1.4
	v1.1.3
	v1.1.2
	v1.1.1
	v1.1.0
)
