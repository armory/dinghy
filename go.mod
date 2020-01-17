module github.com/armory-io/dinghy

require (
	github.com/armory/dinghy v1.0.2-0.20200117042130-7dfd930b1bf2
	github.com/armory/go-yaml-tools v0.0.0-20200117023812-8d4231ac6a03
	github.com/armory/plank v1.2.2-0.20190814000943-6317f9d5d508
	github.com/ghodss/yaml v1.0.0
	github.com/golang/mock v1.3.1
	github.com/hashicorp/go-retryablehttp v0.6.2
	github.com/hashicorp/hcl v1.0.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.3.0
	gopkg.in/yaml.v2 v2.2.2
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

go 1.13
