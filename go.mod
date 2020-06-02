module github.com/armory-io/dinghy

require (
	cloud.google.com/go v0.56.0 // indirect
	github.com/armory/dinghy v1.0.2-0.20200602230508-250dbda98d76
	github.com/armory/go-yaml-tools v0.0.0-20200316192928-75770481ad01
	github.com/armory/plank/v3 v3.1.0
	github.com/aws/aws-sdk-go v1.30.7 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-redis/redis v6.15.7+incompatible // indirect
	github.com/golang/mock v1.4.3
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/hashicorp/go-hclog v0.12.0 // indirect
	github.com/hashicorp/go-multierror v1.1.0 // indirect
	github.com/hashicorp/go-retryablehttp v0.6.5
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/hcl v1.0.0
	// Mergo 0.3.9 fails to override, keep at v0.3.8
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/jinzhu/copier v0.0.0-20190924061706-b57f9002281a // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/mitchellh/mapstructure v1.2.2
	github.com/pierrec/lz4 v2.5.0+incompatible // indirect
	github.com/sirupsen/logrus v1.5.0
	github.com/spf13/afero v1.2.2 // indirect
	github.com/stretchr/testify v1.5.1
	golang.org/x/crypto v0.0.0-20200406173513-056763e48d71 // indirect
	golang.org/x/sys v0.0.0-20200409092240-59c9f1ba88fa // indirect
	golang.org/x/tools v0.0.0-20200410181643-e3f0bd94ad67 // indirect
	google.golang.org/api v0.21.0 // indirect
	google.golang.org/genproto v0.0.0-20200410110633-0848e9f44c36 // indirect
	google.golang.org/grpc v1.28.1 // indirect
	gopkg.in/square/go-jose.v2 v2.4.1 // indirect
	gopkg.in/yaml.v2 v2.2.8
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

go 1.13
