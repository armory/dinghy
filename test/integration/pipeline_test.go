package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/armory-io/dinghy/pkg/cache"
	"github.com/armory-io/dinghy/pkg/dinghyfile"
	"github.com/armory-io/dinghy/pkg/git"
	"github.com/armory-io/dinghy/pkg/git/dummy"
	"github.com/armory-io/dinghy/pkg/git/github"
	"github.com/armory-io/dinghy/pkg/settings"
	"github.com/armory-io/dinghy/pkg/web"
)

const dinghfNew = `{
	"application": "dinghyintegration",
	"pipelines": [
		{
			"application": "dinghyintegration",
			"keepWaitingPipelines": false,
			"limitConcurrent": false,
			"name": "This is new",
			"stages": [
				{{ module "mod1" }},
				{{ module "mod2" }}
			],
			"triggers": []
		}
	]
}`

const dinghfEmpty = `{
	"application": "dinghyintegration",
	"pipelines": []
}`

const mod1 = `{
	"clusters": [
	  {
		"account": "preprod",
		"amiName": "${#stage( 'us-west-2 Build Launch Details' )['context']['IMAGE_ID']}",
		"application": "${#stage( 'us-west-2 Build Launch Details' )['context']['SERVICE']}",
		"associatePublicIpAddress": "${#stage( 'us-west-2 Build Launch Details' )['context']['ASSOCIATE_PUBLIC_IP_ADDRESS']}",
		"availabilityZones": {
		  "${#stage( 'us-west-2 Build Launch Details' )['context']['REGION']}": [
			"${#stage( 'us-west-2 Build Launch Details' )['context']['AVAILABILITY_ZONE_1']}",
			"${#stage( 'us-west-2 Build Launch Details' )['context']['AVAILABILITY_ZONE_2']}",
			"${#stage( 'us-west-2 Build Launch Details' )['context']['AVAILABILITY_ZONE_3']}"
		  ]
		},
		"base64UserData": "${#stage( 'us-west-2 Build Launch Details' )['context']['USER_DATA']}",
		"blockDevices": "${#readJson(#stage( 'us-west-2 Build Launch Details' )['context']['BLOCK_DEVICES'])}",
		"capacity": {
		  "desired": "${#stage( 'us-west-2 Build Launch Details' )['context']['DESIRED_CAPACITY']}",
		  "max": "${#stage( 'us-west-2 Build Launch Details' )['context']['MAX_SIZE']}",
		  "min": "${#stage( 'us-west-2 Build Launch Details' )['context']['MIN_SIZE']}"
		},
		"cloudProvider": "aws",
		"cooldown": 300,
		"copySourceCustomBlockDeviceMappings": true,
		"ebsOptimized": false,
		"enabledMetrics": [],
		"freeFormDetails": "${#stage( 'us-west-2 Build Launch Details' )['context']['VERSION']}",
		"healthCheckGracePeriod": 0,
		"healthCheckType": "${#stage( 'us-west-2 Build Launch Details' )['context']['HEALTH_CHECK_TYPE']}",
		"iamRole": "${#stage( 'us-west-2 Build Launch Details' )['context']['INSTANCE_PROFILE_NAME']}",
		"instanceMonitoring": false,
		"instanceType": "${#stage( 'us-west-2 Build Launch Details' )['context']['INSTANCE_TYPE']}",
		"keyPair": "${#stage( 'us-west-2 Build Launch Details' )['context']['KEY_NAME']}",
		"loadBalancers": [],
		"moniker": {
		  "app": "${#stage( 'us-west-2 Build Launch Details' )['context']['SERVICE']}",
		  "cluster": "${#stage( 'us-west-2 Build Launch Details' )['context']['CLUSTER']}",
		  "detail": "${#stage( 'us-west-2 Build Launch Details' )['context']['VERSION']}",
		  "stack": "${#stage( 'us-west-2 Build Launch Details' )['context']['STACK']}"
		},
		"provider": "aws",
		"securityGroups": "${#readJson(#stage( 'us-west-2 Build Launch Details' )['context']['SECURITY_GROUPS'])}",
		"spotPrice": "",
		"stack": "${#stage( 'us-west-2 Build Launch Details' )['context']['STACK']}",
		"strategy": "",
		"subnetType": "${#stage( 'us-west-2 Build Launch Details' )['context']['SUBNET_TYPE']}",
		"suspendedProcesses": [],
		"tags": {
		  "Environment": "${#stage( 'us-west-2 Build Launch Details' )['context']['ENVIRONMENT']}",
		  "KeyPairFingerprint": "${#stage( 'us-west-2 Build Launch Details' )['context']['KEY_PAIR_FINGERPRINT']}",
		  "LaunchConfiguration": "${#stage( 'us-west-2 Build Launch Details' )['context']['NAME']}",
		  "Name": "${#stage( 'us-west-2 Build Launch Details' )['context']['INSTANCE_NAME']}",
		  "ServiceName": "${#stage( 'us-west-2 Build Launch Details' )['context']['SERVICE']}",
		  "ServiceVersion": "${#stage( 'us-west-2 Build Launch Details' )['context']['VERSION']}",
		  "System": "${#stage( 'us-west-2 Build Launch Details' )['context']['SYSTEM']}"
		},
		"targetGroups": [],
		"targetHealthyDeployPercentage": 100,
		"terminationPolicies": [
		  "OldestInstance"
		],
		"useAmiBlockDeviceMappings": false,
		"useSourceCapacity": false
	  }
	],
	"comments": "Create AWS objects based on the properties file.",
	"completeOtherBranchesThenFail": false,
	"continuePipeline": true,
	"failPipeline": false,
	"name": "us-west-2 Deploy",
	"overrideTimeout": true,
	"restrictExecutionDuringTimeWindow": false,
	"stageTimeoutMs": 1800000,
	"type": "deploy",
	"refId": "deployuswest2",
	"requisiteStageRefIds": ["buildlaunchconfiguswest2"]
  }`

const mod2 = `{
	"comments": "Build properties file containing relevant information to building a launch config and autoscaling group. This info is used it the proceeding step to build the AWS objects.",
	"refId": "buildlaunchconfiguswest2",
	"requisiteStageRefIds" : [],
	"completeOtherBranchesThenFail": false,
	"continuePipeline": true,
	"failPipeline": false,
	"job": "write-lc-properties-file",
	"master": "",
	"name": "us-west-2 Build Launch Details",
	"parameters":
	{
	  "region": "us_west",
	  "service_name": "${parameters.service}",
	  "service_version": "${parameters.version}",
	  "desired_instance_count": "3"
	}, 
		"stageEnabled": {
	  "expression": "${ parameters.deploy_us_west_default == \"True\" || #stage('us-west-2 Deploy Judgement')['status'] == \"SUCCEEDED\" }",
	  "type": "expression"
	},
	"propertyFile": "launchconfig.properties",
	"type": "jenkins"
  }`

// FileService is for working with repositories
type FileService struct {
	Empty bool
}

// Download a file from github.
func (f *FileService) Download(org, repo, file string) (string, error) {
	if file == "dinghyfile" {
		if f.Empty {
			return dinghfEmpty, nil
		}
		return dinghfNew, nil
	} else if file == "mod1" {
		return mod1, nil
	}
	return mod2, nil
}

// EncodeURL returns the git url for a given org, repo, path
func (f *FileService) EncodeURL(org, repo, path string) string {
	return (&github.FileService{}).EncodeURL(org, repo, path)
}

// DecodeURL takes a url and returns the org, repo, path
func (f *FileService) DecodeURL(url string) (org, repo, path string) {
	return (&github.FileService{}).DecodeURL(url)
}

// TestSpinnakerPipelineUpdate tests pipeline update in spinnaker
// even though it mocks out the github part, it talks to spinnaker
// hence it is an integration test and not a unit-test
func TestSpinnakerPipelineUpdate(t *testing.T) {
	builder := &dinghyfile.PipelineBuilder{
		Depman:     cache.NewMemoryCache(),
		Downloader: &FileService{Empty: false},
	}

	push := &dummy.Push{
		FileNames: []string{settings.S.DinghyFilename},
		RepoName:  settings.S.TemplateRepo,
		OrgName:   settings.S.TemplateOrg,
	}

	err := web.ProcessPush(push, builder)
	assert.Equal(t, nil, err)

	builder.Downloader.(*FileService).Empty = true
	err = web.ProcessPush(push, builder)
	assert.Equal(t, nil, err)
}
