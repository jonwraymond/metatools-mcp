module github.com/jonwraymond/metatools-mcp

go 1.24.4

replace github.com/jonwraymond/metatools-mcp => .

replace github.com/jonwraymond/toolfoundation => ../toolfoundation

replace github.com/jonwraymond/tooldiscovery => ../tooldiscovery

replace github.com/jonwraymond/toolexec => ../toolexec

require (
	github.com/caarlos0/env/v11 v11.3.1
	github.com/docker/docker v27.0.3+incompatible
	github.com/go-viper/mapstructure/v2 v2.0.0-alpha.1
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/jonwraymond/tooldiscovery v0.0.0
	github.com/jonwraymond/toolexec v0.1.0
	github.com/jonwraymond/toolfoundation v0.1.0
	github.com/knadh/koanf/parsers/yaml v0.1.0
	github.com/knadh/koanf/providers/env v0.1.0
	github.com/knadh/koanf/providers/file v0.1.0
	github.com/knadh/koanf/providers/structs v0.1.0
	github.com/knadh/koanf/v2 v2.1.0
	github.com/modelcontextprotocol/go-sdk v1.2.0
	github.com/redis/go-redis/v9 v9.17.3
	github.com/spf13/cobra v1.8.1
	github.com/stretchr/testify v1.11.1
	github.com/tetratelabs/wazero v1.11.0
	golang.org/x/sync v0.10.0
	golang.org/x/time v0.14.0
)

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/RoaringBitmap/roaring/v2 v2.4.5 // indirect
	github.com/bits-and-blooms/bitset v1.22.0 // indirect
	github.com/blevesearch/bleve/v2 v2.5.7 // indirect
	github.com/blevesearch/bleve_index_api v1.2.11 // indirect
	github.com/blevesearch/geo v0.2.4 // indirect
	github.com/blevesearch/go-faiss v1.0.26 // indirect
	github.com/blevesearch/go-porterstemmer v1.0.3 // indirect
	github.com/blevesearch/gtreap v0.1.1 // indirect
	github.com/blevesearch/mmap-go v1.0.4 // indirect
	github.com/blevesearch/scorch_segment_api/v2 v2.3.13 // indirect
	github.com/blevesearch/segment v0.9.1 // indirect
	github.com/blevesearch/snowballstem v0.9.0 // indirect
	github.com/blevesearch/upsidedown_store_api v1.0.2 // indirect
	github.com/blevesearch/vellum v1.1.0 // indirect
	github.com/blevesearch/zapx/v11 v11.4.2 // indirect
	github.com/blevesearch/zapx/v12 v12.4.2 // indirect
	github.com/blevesearch/zapx/v13 v13.4.2 // indirect
	github.com/blevesearch/zapx/v14 v14.4.2 // indirect
	github.com/blevesearch/zapx/v15 v15.4.2 // indirect
	github.com/blevesearch/zapx/v16 v16.2.8 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/jsonschema-go v0.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/json-iterator/go v0.0.0-20171115153421-f7279a603ede // indirect
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/morikuni/aec v1.1.0 // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	go.etcd.io/bbolt v1.4.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.64.0 // indirect
	go.opentelemetry.io/otel v1.39.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.39.0 // indirect
	go.opentelemetry.io/otel/metric v1.39.0 // indirect
	go.opentelemetry.io/otel/trace v1.39.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gotest.tools/v3 v3.5.2 // indirect
)
