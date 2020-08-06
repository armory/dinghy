module github.com/armory/dinghy

require (
	github.com/armory-io/monitoring v0.0.7
	github.com/armory/go-yaml-tools v0.0.0-20200603151141-b037d3988c49
	github.com/armory/plank/v3 v3.4.1
	github.com/go-redis/redis v6.14.1+incompatible
	github.com/golang/mock v1.3.1
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/google/go-github v17.0.0+incompatible
	github.com/gorilla/context v1.1.1 // indirect
	github.com/gorilla/mux v1.6.2
	github.com/hashicorp/go-cleanhttp v0.5.1
	github.com/hashicorp/go-retryablehttp v0.6.2
	github.com/imdario/mergo v0.3.6
	github.com/jinzhu/copier v0.0.0-20180308034124-7e38e58719c3
	github.com/mitchellh/mapstructure v1.1.2
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/prometheus/client_golang v1.7.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/afero v1.1.2 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/xanzy/go-gitlab v0.20.1
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e // indirect
	golang.org/x/sys v0.0.0-20200615200032-f1bc736245b1 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.2.5 // indirect
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

go 1.13
