/*
* Copyright 2019 Armory, Inc.

* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at

*    http://www.apache.org/licenses/LICENSE-2.0

* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package dinghyfile

import (
	"bytes"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/armory/plank"
)

// Test the high-level runthrough of ProcessDinghyfile
func TestProcessDinghyfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rendered := `{"application":"biff"}`

	renderer := NewMockRenderer(ctrl)
	renderer.EXPECT().Render(gomock.Eq("myorg"), gomock.Eq("myrepo"), gomock.Eq("the/full/path"), gomock.Any()).Return(bytes.NewBuffer([]byte(rendered)), nil).Times(1)
	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetApplication(gomock.Eq("biff")).Return(&plank.Application{}, nil).Times(1)
	client.EXPECT().GetPipelines(gomock.Eq("biff")).Return([]plank.Pipeline{}, nil).Times(1)
	// Never gets UpsertPipeline because our Render returns no pipelines.
	client.EXPECT().UpsertPipeline(gomock.Any(), "").Return(nil).Times(0)
	pb := PipelineBuilder{
		Renderer: renderer,
		Client:   client,
	}
	assert.Nil(t, pb.ProcessDinghyfile("myorg", "myrepo", "the/full/path"))
}

// TestUpdateDinghyfile ONLY tests the function "updateDinghyfile" which,
// despite its name, doesn't really update anything, it just unmarshals
// the payload into the Dinghyfile{} struct.
func TestUpdateDinghyfile(t *testing.T) {

	cases := map[string]struct {
		dinghyfile []byte
		spec       plank.Application
	}{
		"no_spec": {
			dinghyfile: []byte(`{
				"application": "application"
			}`),
			spec: plank.Application{
				Name:  "application",
				Email: DefaultEmail,
				DataSources: plank.DataSourcesType{
					Enabled:  []string{},
					Disabled: []string{},
				},
			},
		},
		"spec": {
			dinghyfile: []byte(`{
				"application": "application",
				"spec": {
					"name": "specname"
				}
			}`),
			spec: plank.Application{
				Name:  "specname",
				Email: DefaultEmail,
				DataSources: plank.DataSourcesType{
					Enabled:  []string{},
					Disabled: []string{},
				},
			},
		},
		"spec_with_email": {
			dinghyfile: []byte(`{
				"application": "application",
				"spec": {
					"name": "specname",
					"email": "somebody@email.com"
				}
			}`),
			spec: plank.Application{
				Name:  "specname",
				Email: "somebody@email.com",
				DataSources: plank.DataSourcesType{
					Enabled:  []string{},
					Disabled: []string{},
				},
			},
		},
		"spec_with_canary": {
			dinghyfile: []byte(`{
				"application": "application",
				"spec": {
					"name": "specname",
					"email": "somebody@email.com",
					"dataSources": {
						"disabled":[],
						"enabled":["canaryConfigs"]
					}
				}
			}`),
			spec: plank.Application{
				Name:  "specname",
				Email: "somebody@email.com",
				DataSources: plank.DataSourcesType{
					Enabled:  []string{"canaryConfigs"},
					Disabled: []string{},
				},
			},
		},
	}

	for testName, c := range cases {
		t.Run(testName, func(t *testing.T) {
			d, _ := UpdateDinghyfile(c.dinghyfile)
			assert.Equal(t, d.ApplicationSpec, c.spec)
		})
	}

	fullCases := map[string]struct {
		dinghyRaw    []byte
		dinghyStruct Dinghyfile
	}{
		"dinghyraw_with_expectedArtifacts": {
			dinghyRaw: []byte(`{
				"application": "foo",
				"spec": {
					"name": "foo",
					"email": "foo@test.com",
					"dataSources": {
						"disabled":[],
						"enabled":[]
					}
				},
				"pipelines": [
					{
						"name": "test",
						"expectedArtifacts": [
							{
								"foo": {
									"bar": "baz"
								}
							}
						],
						"stages": [
							{
								"foo": {
									"bar": "baz"
								}
							}
						]
					}
				]
			}`),
			dinghyStruct: Dinghyfile{
				Application: "foo",
				ApplicationSpec: plank.Application{
					Name:  "foo",
					Email: "foo@test.com",
					DataSources: plank.DataSourcesType{
						Enabled:  []string{},
						Disabled: []string{},
					},
				},
				Pipelines: []plank.Pipeline{
					{
						Name: "test",
						ExpectedArtifacts: []map[string]interface{}{
							{
								"foo": map[string]interface{}{
									"bar": "baz",
								},
							},
						},
						Stages: []map[string]interface{}{
							{
								"foo": map[string]interface{}{
									"bar": "baz",
								},
							},
						},
					},
				},
			},
		},
	}

	for testName, c := range fullCases {
		t.Run(testName, func(t *testing.T) {
			d, _ := UpdateDinghyfile(c.dinghyRaw)
			assert.Equal(t, d, c.dinghyStruct)
		})
	}
}
