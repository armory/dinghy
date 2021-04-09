package yeti

import (
	"bytes"
	"github.com/armory-io/dinghy/utils/mocks"
	ossSettings "github.com/armory/dinghy/pkg/settings/global"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
)

func init() {
	HttpClient = &mocks.MockClient{}
}

func TestClient_GetSettings(t *testing.T) {
	type fields struct {
		YetiUrl     string
		ArmoryOrgId string
	}
	type args struct {
		environmentId  string
		organizationID string
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		responseJson string
		want         *ossSettings.Settings
		wantErr      bool
	}{
		{
			name: "Dinghy settings should be created successfully",
			args: args{
				environmentId:  "6604133f-f21a-4d38-8d76-b5923874cfd9",
				organizationID: "0043dba7-200b-4bea-857b-a1b060089197",
			},
			responseJson: `{
				"id": "7ccc9645-2b0e-448a-8f9a-973f824988a6",
				"orgId": "0043dba7-200b-4bea-857b-a1b060089197",
				"envId": "6604133f-f21a-4d38-8d76-b5923874cfd9",
				"resourceType": "ARMORY",
				"resourceKind": "dinghy",
				"createdTs": "2021-03-17T20:36:18Z",
				"updatedTs": "2021-03-24T06:39:26Z",
				"dinghy": {
					"LogEventTTLMinutes": 0,
					"deck": {},
					"echo": {},
					"fiat": {},
					"front50": {},
					"http": {
						"CacertFile": "",
						"ClientCertFile": "",
						"ClientKeyFile": "",
						"ClientKeyPassword": ""
					},
					"logging": {
						"remote": {
							"customerId": "",
							"enabled": false,
							"endpoint": "",
							"version": ""
						}
					},
					"orca": {},
					"redis": {},
					"secrets": {
						"vault": {
							"Token": "",
							"authMethod": "",
							"enabled": false,
							"namespace": "",
							"password": "",
							"path": "",
							"role": "",
							"url": "",
							"userAuthPath": "",
							"username": ""
						}
					},
					"server": {
						"Host": "",
						"Port": 0,
						"Ssl": {
							"CAcertFile": "",
							"CertFile": "",
							"ClientAuth": "",
							"Enabled": false,
							"KeyFile": "",
							"KeyPassword": ""
						}
					},
					"sql": {
						"baseUrl": "",
						"databaseName": "",
						"eventlogsOnly": false,
						"password": "",
						"user": ""
					},
					"templateOrg": "testUpdate"
				}
			}`,
			want: &ossSettings.Settings{
				TemplateOrg: "testUpdate",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			y := &Client{
				YetiUrl: tt.fields.YetiUrl,
			}
			r := ioutil.NopCloser(bytes.NewReader([]byte(tt.responseJson)))
			mocks.GetDoFunc = func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       r,
				}, nil
			}
			got, err := y.GetSettings(tt.args.environmentId, tt.args.organizationID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSettings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSettings() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewYetiClient(t *testing.T) {
	type args struct {
		endpoint    string
		armoryOrgId string
	}
	tests := []struct {
		name string
		args args
		want Client
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewYetiClient(tt.args.endpoint); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewYetiClient() = %v, want %v", got, tt.want)
			}
		})
	}
}
