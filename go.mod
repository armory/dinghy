module github.com/armory/dinghy

require (
	github.com/Masterminds/sprig/v3 v3.1.0
	github.com/armory-io/monitoring v0.0.7
	github.com/armory/go-yaml-tools v0.0.2
	github.com/armory/plank/v4 v4.0.0-20210507183839-114b4edf7951 // indirect
	github.com/go-redis/redis v6.14.1+incompatible
	github.com/golang/mock v1.3.1
	github.com/google/go-github/v33 v33.0.0
	github.com/google/uuid v1.1.1
	github.com/gorilla/context v1.1.1 // indirect
	github.com/gorilla/mux v1.6.2
	github.com/hashicorp/go-cleanhttp v0.5.1
	github.com/hashicorp/go-retryablehttp v0.6.2
	github.com/imdario/mergo v0.3.8
	github.com/jinzhu/copier v0.0.0-20180308034124-7e38e58719c3
	github.com/mitchellh/mapstructure v1.1.2
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/otiai10/copy v1.5.0
	github.com/prometheus/client_golang v1.7.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/afero v1.1.2 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/xanzy/go-gitlab v0.20.1
	go.opentelemetry.io/otel v0.16.0
	go.opentelemetry.io/otel/exporters/stdout v0.16.0
	go.opentelemetry.io/otel/sdk v0.16.0
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sys v0.0.0-20210315160823-c6e025ad8005 // indirect
	golang.org/x/tools v0.1.0 // indirect
	gorm.io/driver/mysql v1.0.3
	gorm.io/gorm v1.20.7
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

go 1.13
