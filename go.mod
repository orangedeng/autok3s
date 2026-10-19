module github.com/orangedeng/autok3s

go 1.26.3

replace (
	go4.org/unsafe/assume-no-moving-gc => go4.org/unsafe/assume-no-moving-gc v0.0.0-20230525183740-e7c30c78aeb2
	inet.af/netaddr => inet.af/netaddr v0.0.0-20230525184311-b8eac61e914a
)

require (
	github.com/AlecAivazis/survey/v2 v2.3.5
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig/v3 v3.2.3
	github.com/alexellis/go-execute v0.0.0-20200124154445-8697e4e28c5e
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.381
	github.com/creack/pty v1.1.24
	github.com/docker/docker v28.5.2+incompatible
	github.com/docker/go-units v0.5.0
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/glebarez/sqlite v1.4.3
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674
	github.com/hpcloud/tail v1.0.0
	github.com/imdario/mergo v0.3.14 // indirect
	github.com/k3d-io/k3d/v5 v5.9.0
	github.com/moby/term v0.5.2
	github.com/morikuni/aec v1.1.0
	github.com/olekukonko/tablewriter v0.0.5
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.35.1
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8
	github.com/pkg/errors v0.9.1
	github.com/pkg/sftp v1.13.6
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/rancher/apiserver v0.0.0-20240207153957-4fd7d821d952
	github.com/rancher/wharfie v0.7.0
	github.com/rancher/wrangler/v2 v2.1.3
	github.com/sirupsen/logrus v1.9.4
	github.com/spf13/cobra v1.10.2
	github.com/spf13/pflag v1.0.10
	github.com/stretchr/testify v1.11.1
	github.com/tencentcloud/tencentcloud-sdk-go v3.0.233+incompatible
	golang.org/x/crypto v0.51.0
	golang.org/x/net v0.55.0
	golang.org/x/oauth2 v0.36.0
	golang.org/x/sync v0.20.0
	golang.org/x/term v0.43.0
	google.golang.org/api v0.153.0
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/gorm v1.23.4
	helm.sh/helm/v3 v3.14.4
	k8s.io/api v0.36.1
	k8s.io/apimachinery v0.36.1
	k8s.io/client-go v0.36.1
)

require (
	github.com/aws/aws-sdk-go-v2 v1.47.3
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.339.0
	github.com/aws/smithy-go v1.28.5
	github.com/moby/sys/signal v0.7.0
	k8s.io/kube-openapi v0.0.0-20260520065146-aa012df4f4af
	sigs.k8s.io/yaml v1.6.0
)

require (
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	dario.cat/mergo v1.0.2 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.2.1 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.5.6 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.8.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.21 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.14.6 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/stargz-snapshotter/estargz v0.16.3 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dimchansky/utfbom v1.1.1 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/cli v28.5.2+incompatible // indirect
	github.com/docker/distribution v2.8.3+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.9.7 // indirect
	github.com/docker/go v1.5.1-1.0.20160303222718-d30aec9fd63c // indirect
	github.com/docker/go-connections v0.7.0 // indirect
	github.com/emicklei/go-restful/v3 v3.13.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/fvbommel/sortorder v1.1.0 // indirect
	github.com/fxamacker/cbor/v2 v2.9.2 // indirect
	github.com/glebarez/go-sqlite v1.16.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/swag v0.25.4 // indirect
	github.com/go-openapi/swag/cmdutils v0.25.4 // indirect
	github.com/go-openapi/swag/conv v0.25.4 // indirect
	github.com/go-openapi/swag/fileutils v0.25.4 // indirect
	github.com/go-openapi/swag/jsonname v0.25.4 // indirect
	github.com/go-openapi/swag/jsonutils v0.25.4 // indirect
	github.com/go-openapi/swag/loading v0.25.4 // indirect
	github.com/go-openapi/swag/mangling v0.25.4 // indirect
	github.com/go-openapi/swag/netutils v0.25.4 // indirect
	github.com/go-openapi/swag/stringutils v0.25.4 // indirect
	github.com/go-openapi/swag/typeutils v0.25.4 // indirect
	github.com/go-openapi/swag/yamlutils v0.25.4 // indirect
	github.com/go-test/deep v1.1.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/goodhosts/hostsfile v0.1.7 // indirect
	github.com/google/gnostic-models v0.7.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-containerregistry v0.20.6 // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.12.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.29.0 // indirect
	github.com/huandu/xstrings v1.4.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.4 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/klauspost/compress v1.18.6 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/magefile/mage v1.17.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/mattn/go-runewidth v0.0.24 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/miekg/pkcs11 v1.1.1 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/go-archive v0.1.0 // indirect
	github.com/moby/patternmatcher v0.6.1 // indirect
	github.com/moby/sys/atomicwriter v0.1.0 // indirect
	github.com/moby/sys/sequential v0.6.0 // indirect
	github.com/moby/sys/user v0.4.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/pelletier/go-toml/v2 v2.3.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20200410134404-eec4a21b6bb0 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/viper v1.21.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/vbatts/tar-split v0.12.1 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.69.0 // indirect
	go.opentelemetry.io/otel v1.44.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.44.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.44.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.44.0 // indirect
	go.opentelemetry.io/otel/metric v1.44.0 // indirect
	go.opentelemetry.io/otel/sdk v1.44.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.44.0 // indirect
	go.opentelemetry.io/otel/trace v1.44.0 // indirect
	go.opentelemetry.io/proto/otlp v1.10.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	go4.org/netipx v0.0.0-20231129151722-fdeea329fbba // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/grpc v1.81.1 // indirect
	google.golang.org/protobuf v1.36.12-0.20260120151049-f2248ac996af // indirect
	gopkg.in/evanphx/json-patch.v4 v4.13.0 // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/apiserver v0.34.2 // indirect
	k8s.io/component-base v0.34.2 // indirect
	k8s.io/klog/v2 v2.140.0 // indirect
	k8s.io/streaming v0.36.1 // indirect
	k8s.io/utils v0.0.0-20260507154919-ff6756f316d2 // indirect
	modernc.org/libc v1.14.12 // indirect
	modernc.org/mathutil v1.4.1 // indirect
	modernc.org/memory v1.0.7 // indirect
	modernc.org/sqlite v1.16.0 // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.4.0 // indirect
)
