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
	"errors"
	"github.com/armory/dinghy/pkg/dinghyfile/pipebuilder"
	"github.com/armory/dinghy/pkg/events"
	"github.com/armory/dinghy/pkg/log"
	"github.com/armory/dinghy/pkg/util"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/assert"

	"github.com/armory/plank/v4"

	"github.com/armory/dinghy/pkg/mock"
	"github.com/armory/dinghy/pkg/notifiers"
)

// Test the high-level runthrough of ProcessDinghyfile
func TestProcessDinghyfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rendered := `{"application":"biff"}`

	renderer := NewMockParser(ctrl)
	renderer.EXPECT().Parse(gomock.Eq("myorg"), gomock.Eq("myrepo"), gomock.Eq("the/full/path"), gomock.Eq("mybranch"), gomock.Any()).Return(bytes.NewBuffer([]byte(rendered)), nil).Times(1)

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetApplication(gomock.Eq("biff"), "").Return(&plank.Application{}, nil).Times(1)
	client.EXPECT().GetPipelines(gomock.Eq("biff"), "").Return([]plank.Pipeline{}, nil).Times(1)

	logger := mock.NewMockDinghyLog(ctrl)
	logger.EXPECT().Infof(gomock.Eq("Unmarshalled: %v"), gomock.Any()).Times(1)
	logger.EXPECT().Infof(gomock.Eq("Found pipelines for %v: %v"), gomock.Any()).Times(1)
	logger.EXPECT().Infof(gomock.Eq("Dinghyfile struct: %v"), gomock.Any()).Times(1)
	logger.EXPECT().Infof(gomock.Eq("Updated: %s"), gomock.Any()).Times(1)
	logger.EXPECT().Infof(gomock.Eq("Compiled: %s"), gomock.Any()).Times(1)
	logger.EXPECT().Info(gomock.Eq("Looking up existing pipelines")).Times(1)
	logger.EXPECT().Info(gomock.Eq("Validations for stage refs were successful")).Times(1)
	logger.EXPECT().Info(gomock.Eq("Validations for app notifications were successful")).Times(1)

	// Because we've set the renderer, we should NOT get this message...
	logger.EXPECT().Info(gomock.Eq("Calling DetermineRenderer")).Times(0)

	// Never gets UpsertPipeline because our Render returns no pipelines.
	client.EXPECT().UpsertPipeline(gomock.Any(), "", "").Return(nil).Times(0)
	pb := testPipelineBuilder()
	pb.Parser = renderer
	pb.Client = client
	pb.Logger = logger
	if _, err := pb.ProcessDinghyfile("myorg", "myrepo", "the/full/path", "mybranch"); err != nil {
		t.Fail()
	}
}

func TestProcessDinghyfileValidate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rendered := `{"application":"biff"}`

	renderer := NewMockParser(ctrl)
	renderer.EXPECT().Parse(gomock.Eq("myorg"), gomock.Eq("myrepo"), gomock.Eq("the/full/path"), gomock.Eq("mybranch"), gomock.Any()).Return(bytes.NewBuffer([]byte(rendered)), nil).Times(1)

	client := NewMockPlankClient(ctrl)

	logger := mock.NewMockDinghyLog(ctrl)
	logger.EXPECT().Infof(gomock.Eq("Unmarshalled: %v"), gomock.Any()).Times(1)
	logger.EXPECT().Infof(gomock.Eq("Dinghyfile struct: %v"), gomock.Any()).Times(1)
	logger.EXPECT().Infof(gomock.Eq("Updated: %s"), gomock.Any()).Times(1)
	logger.EXPECT().Infof(gomock.Eq("Compiled: %s"), gomock.Any()).Times(1)
	logger.EXPECT().Info(gomock.Eq("Validations for stage refs were successful")).Times(1)
	logger.EXPECT().Info(gomock.Eq("Validations for app notifications were successful")).Times(1)
	logger.EXPECT().Info(gomock.Eq("Validation finished successfully")).Times(1)

	// Because we've set the renderer, we should NOT get this message...
	logger.EXPECT().Info(gomock.Eq("Calling DetermineRenderer")).Times(0)

	pb := testPipelineBuilder()
	pb.Action = pipebuilder.Validate
	pb.Parser = renderer
	pb.Client = client
	pb.Logger = logger

	if _, err := pb.ProcessDinghyfile("myorg", "myrepo", "the/full/path", "mybranch"); err != nil {
		t.Fail()
	}
}

// Test updateapp
func TestUpdateApplication(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	existingPipeline := plank.Pipeline{Name: "ExistingPipeline", ID: "ExistingID", Application: "testapp", Locked: &plank.PipelineLockType{true, true}}
	deletedPipeline := plank.Pipeline{Name: "DeletedPipeline", ID: "DeletedID", Application: "testapp"}
	newPipeline := plank.Pipeline{Name: "NewPipeline", ID: "NewID", Locked: &plank.PipelineLockType{true, true}}

	existing := []plank.Pipeline{existingPipeline, deletedPipeline}
	newPipelines := []plank.Pipeline{existingPipeline, newPipeline}

	testapp := &plank.Application{Name: "testapp"}

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetApplication("testapp", "").Return(nil, nil).Times(1)
	client.EXPECT().GetPipelines(gomock.Eq("testapp"), "").Return(existing, nil).Times(1)
	client.EXPECT().UpsertPipeline(gomock.Eq(existingPipeline), gomock.Eq(existingPipeline.ID), "").Return(nil).Times(1)
	client.EXPECT().UpsertPipeline(gomock.Eq(newPipeline), gomock.Eq(newPipeline.ID), "").Return(errors.New("upsert fail test")).Times(1)
	// Should not get called at all, because of earlier error
	client.EXPECT().DeletePipeline(gomock.Eq(deletedPipeline), "").Times(0)
	client.EXPECT().UpdateApplication(gomock.Eq(*testapp), "").Return(nil).Times(1)
	client.EXPECT().UpdateApplicationNotifications(gomock.Eq(testapp.Notifications), gomock.Eq(testapp.Name), "").Return(nil).Times(1)

	b := testPipelineBuilder()
	var gvMap = make(map[string]interface{})
	gvMap["save_app_on_update"] = true
	b.GlobalVariablesMap = gvMap
	b.Client = client

	err := b.updatePipelines(testapp, newPipelines, true, "true")
	assert.NotNil(t, err)
	assert.Equal(t, "upsert fail test", err.Error())
}

// Note: This ALSO tests the error case where the renderer fails (in this
// example, because the rendered file path info is invalid)
func TestProcessDinghyfileDefaultRenderer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockDinghyLog(ctrl)
	logger.EXPECT().Info("Calling DetermineParser").Times(1)
	logger.EXPECT().Errorf(gomock.Eq("Failed to download %s/%s/%s/%s"), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	logger.EXPECT().Errorf(gomock.Eq("Failed to parse dinghyfile %s: %s"), gomock.Eq("notfound"), gomock.Eq("File not found")).Times(1)

	pb := testPipelineBuilder()
	pb.Logger = logger
	_, res := pb.ProcessDinghyfile("fake", "news", "notfound", "branch")
	assert.NotNil(t, res)
}

func TestProcessDinghyfileFailedUnmarshal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rendered := `{blargh}`

	renderer := NewMockParser(ctrl)
	renderer.EXPECT().Parse(gomock.Eq("myorg"), gomock.Eq("myrepo"), gomock.Eq("the/full/path"), gomock.Eq("mybranch"), gomock.Any()).Return(bytes.NewBuffer([]byte(rendered)), nil).Times(1)

	logger := mock.NewMockDinghyLog(ctrl)
	logger.EXPECT().Warnf(gomock.Eq("UpdateDinghyfile malformed syntax: %s"), gomock.Any()).Times(1)
	logger.EXPECT().Errorf(gomock.Eq("Failed to update dinghyfile %s: %s"), gomock.Any()).Times(1)
	logger.EXPECT().Errorf(gomock.Eq("update-dinghyfile-unmarshal-err: %s"), gomock.Any()).Times(1)
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	pb := testPipelineBuilder()
	pb.Logger = logger
	pb.Parser = renderer
	dinghyfileParsed, res := pb.ProcessDinghyfile("myorg", "myrepo", "the/full/path", "mybranch")
	assert.Equal(t, rendered, dinghyfileParsed)
	assert.NotNil(t, res)
}

func TestProcessDinghyfileFailedUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rendered := `{"application": "testapp"}`

	renderer := NewMockParser(ctrl)
	renderer.EXPECT().Parse(gomock.Eq("myorg"), gomock.Eq("myrepo"), gomock.Eq("the/full/path"), gomock.Eq("mybranch"), gomock.Any()).Return(bytes.NewBuffer([]byte(rendered)), nil).Times(1)

	logger := mock.NewMockDinghyLog(ctrl)
	logger.EXPECT().Infof(gomock.Eq("Creating application '%s'..."), gomock.Eq("testapp")).Times(1)
	logger.EXPECT().Errorf("Failed to create application (%s)", gomock.Any())
	logger.EXPECT().Errorf(gomock.Eq("Failed to update Pipelines for %s: %s"), gomock.Eq("the/full/path")).Times(1)
	logger.EXPECT().Info(gomock.Eq("Validations for stage refs were successful")).Times(1)
	logger.EXPECT().Info(gomock.Eq("Validations for app notifications were successful")).Times(1)
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetApplication(gomock.Eq("testapp"), "").Return(nil, &plank.FailedResponse{StatusCode: 404}).Times(1)
	client.EXPECT().CreateApplication(gomock.Any(), "").Return(errors.New("boom")).Times(1)
	client.EXPECT().GetApplicationNotifications(gomock.Eq("testapp"), "").Return(nil, &plank.FailedResponse{StatusCode: 404}).Times(1)

	pb := testPipelineBuilder()
	pb.Logger = logger
	pb.Parser = renderer
	pb.Client = client
	dinghyfileParsed, res := pb.ProcessDinghyfile("myorg", "myrepo", "the/full/path", "mybranch")
	assert.Equal(t, rendered, dinghyfileParsed)
	assert.NotNil(t, res)
	assert.Equal(t, "boom", res.Error())
}

func TestProcessDinghyfileFailedValidation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rendered := `{
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
								"failPipeline": true,
								"judgmentInputs": [],
								"name": "Manual Judgment 2",
								"notifications": [],
								"refId": "mj2",
								"type": "manualJudgment"
							},
							{
								"failPipeline": true,
								"judgmentInputs": [],
								"name": "Manual Judgment 2",
								"notifications": [],
								"refId": "mj2",
								"type": "manualJudgment"
							}
						]
					}
				]
			}`

	renderer := NewMockParser(ctrl)
	renderer.EXPECT().Parse(gomock.Eq("myorg"), gomock.Eq("myrepo"), gomock.Eq("the/full/path"), gomock.Eq("mybranch"), gomock.Any()).Return(bytes.NewBuffer([]byte(rendered)), nil).Times(1)

	logger := mock.NewMockDinghyLog(ctrl)
	logger.EXPECT().Errorf(gomock.Eq("Failed to validate stage refs for pipeline: %s"), gomock.Any()).Times(1)
	logger.EXPECT().Errorf(gomock.Eq("validate-pipelines-stagerefs-err: %s"), rendered)
	logger.EXPECT().Errorf(gomock.Eq("Failed to validate pipelines %s"), gomock.Eq("the/full/path")).Times(1)
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetApplicationNotifications(gomock.Eq("foo"), "").Return(nil, &plank.FailedResponse{StatusCode: 404}).Times(1)

	pb := testPipelineBuilder()
	pb.Logger = logger
	pb.Parser = renderer
	pb.Client = client
	dinghyfileParsed, res := pb.ProcessDinghyfile("myorg", "myrepo", "the/full/path", "mybranch")
	assert.Equal(t, rendered, dinghyfileParsed)
	assert.NotNil(t, res)
	assert.Equal(t, "Duplicate stage refId mj2 field found", res.Error())
}

// TestUpdateDinghyfile ONLY tests the function "updateDinghyfile" which,
// despite its name, doesn't really update anything, it just unmarshals
// the payload into the Dinghyfile{} struct.
func TestUpdateDinghyfile(t *testing.T) {
	b := testPipelineBuilder()

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
				DataSources: &plank.DataSourcesType{
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
				DataSources: &plank.DataSourcesType{
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
				DataSources: &plank.DataSourcesType{
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
				DataSources: &plank.DataSourcesType{
					Enabled:  []string{"canaryConfigs"},
					Disabled: []string{},
				},
			},
		},
	}

	for testName, c := range cases {
		t.Run(testName, func(t *testing.T) {
			d, _ := b.UpdateDinghyfile(c.dinghyfile)
			assert.Equal(t, c.spec, d.ApplicationSpec)
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
					DataSources: &plank.DataSourcesType{
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
			d, _ := b.UpdateDinghyfile(c.dinghyRaw)
			assert.Equal(t, d, c.dinghyStruct)
		})
	}
}

func TestValidatePipelines(t *testing.T) {
	b := testPipelineBuilder()

	fullCases := map[string]struct {
		dinghyRaw []byte
		result    error
	}{
		"dinghyraw_fail_no_refids": {
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
								"failPipeline": true,
								"judgmentInputs": [],
								"name": "Manual Judgment 2",
								"notifications": [],
								"type": "manualJudgment"
							}
						]
					}
				]
			}`),
			result: nil,
		},
		"dinghyraw_fail_circular_reference": {
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
								"failPipeline": true,
								"judgmentInputs": [],
								"name": "Manual Judgment 2",
								"notifications": [],
								"refId": "mj2",
								"requisiteStageRefIds": [
									"mj2"
								],
								"type": "manualJudgment"
							}
						]
					}
				]
			}`),
			result: errors.New("mj2 refers to itself. Circular references are not supported"),
		},
		"dinghyraw_passes_no_stages": {
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
						]
					}
				]
			}`),
			result: nil,
		},
		"dinghyraw_passes_all_stages_in_order": {
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
								"failPipeline": true,
								"judgmentInputs": [],
								"name": "Manual Judgment 2",
								"notifications": [],
								"refId": "1",
								"requisiteStageRefIds": [
								],
								"type": "manualJudgment"
							},
							{
								"failPipeline": true,
								"judgmentInputs": [],
								"name": "Manual Judgment 2",
								"notifications": [],
								"refId": "2",
								"requisiteStageRefIds": [
									"1"
								],
								"type": "manualJudgment"
							},
							{
								"failPipeline": true,
								"judgmentInputs": [],
								"name": "Manual Judgment 2",
								"notifications": [],
								"refId": "mj2",
								"requisiteStageRefIds": [
									"2","1"
								],
								"type": "manualJudgment"
							}
						]
					}
				]
			}`),
			result: nil,
		},
		"dinghyraw_fails_duplicated": {
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
								"failPipeline": true,
								"judgmentInputs": [],
								"name": "Manual Judgment 2",
								"notifications": [],
								"refId": "mj2",
								"type": "manualJudgment"
							},
							{
								"failPipeline": true,
								"judgmentInputs": [],
								"name": "Manual Judgment 2",
								"notifications": [],
								"refId": "mj2",
								"type": "manualJudgment"
							}
						]
					}
				]
			}`),
			result: errors.New("Duplicate stage refId mj2 field found"),
		},
	}

	for testName, c := range fullCases {
		t.Run(testName, func(t *testing.T) {
			d := NewDinghyfile()
			// try every parser, maybe we'll get lucky
			parseErrs := 0
			for _, ums := range b.Ums {
				if err := ums.Unmarshal(c.dinghyRaw, &d); err != nil {
					b.Logger.Error("Dinghyfile malformed syntax: %s", err.Error())
					parseErrs++
					continue
				}
			}
			//Log parsing issues as an error in test
			if parseErrs != 0 {
				assert.True(t, true, false)
			} else {
				err := b.ValidatePipelines(d, c.dinghyRaw)
				assert.Equal(t, c.result, err)
			}
		})
	}
}

func TestValidateAppNotifications(t *testing.T) {
	b := testPipelineBuilder()

	fullCases := map[string]struct {
		dinghyRaw []byte
		result    error
	}{
		"dinghyraw_pass_empty": {
			dinghyRaw: []byte(`{
				"application": "foo",
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
								"failPipeline": true,
								"judgmentInputs": [],
								"name": "Manual Judgment 2",
								"notifications": [],
								"type": "manualJudgment"
							}
						]
					}
				]
			}`),
			result: nil,
		},
		"dinghyraw_fail_no_slice_notif": {
			dinghyRaw: []byte(`{
				"application": "foo",
				"spec": {
					"name": "foo",
					"email": "foo@test.com",
					"dataSources": {
						"disabled":[],
						"enabled":[]
					},
					"notifications" : {
						"slack": [{
							"when": [ "pipeline.complete", "pipeline.failed" ],
							"address": "slack-channel"
						}],
						"email": {
							"when": [ "pipeline.complete", "pipeline.failed" ],
							"address": "email-address"
						}
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
								"failPipeline": true,
								"judgmentInputs": [],
								"name": "Manual Judgment 2",
								"notifications": [],
								"refId": "mj2",
								"requisiteStageRefIds": [
									"mj2"
								],
								"type": "manualJudgment"
							}
						]
					}
				]
			}`),
			result: errors.New("application notifications format is invalid for email"),
		},
		"dinghyraw_passes_arrays": {
			dinghyRaw: []byte(`{
				"application": "foo",
				"spec": {
					"name": "foo",
					"email": "foo@test.com",
					"dataSources": {
						"disabled":[],
						"enabled":[]
					},
					"notifications" : {
						"slack": [{
							"when": [ "pipeline.complete", "pipeline.failed" ],
							"address": "slack-channel"
						}],
						"email": [{
							"when": [ "pipeline.complete", "pipeline.failed" ],
							"address": "email-address"
						}]
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
						]
					}
				]
			}`),
			result: nil,
		},
	}

	for testName, c := range fullCases {
		t.Run(testName, func(t *testing.T) {
			d := NewDinghyfile()
			// try every parser, maybe we'll get lucky
			parseErrs := 0
			for _, ums := range b.Ums {
				if err := ums.Unmarshal(c.dinghyRaw, &d); err != nil {
					b.Logger.Error("Dinghyfile malformed syntax: %s", err.Error())
					parseErrs++
					continue
				}
			}
			//Log parsing issues as an error in test
			if parseErrs != 0 {
				assert.True(t, true, false)
			} else {
				err := b.ValidateAppNotifications(d, c.dinghyRaw)
				assert.Equal(t, c.result, err)
			}
		})
	}
}

func TestUpdateDinghyfileMalformed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	b := testPipelineBuilder()
	logger := mock.NewMockDinghyLog(ctrl)
	b.Logger = logger
	logger.EXPECT().Warnf(gomock.Eq("UpdateDinghyfile malformed syntax: %s"), gomock.Any()).Times(1)
	logger.EXPECT().Errorf(gomock.Eq("update-dinghyfile-unmarshal-err: %s"), gomock.Any()).Times(1)

	df, err := b.UpdateDinghyfile([]byte("{ garbage"))
	assert.Equal(t, ErrMalformedJSON, err)
	assert.NotNil(t, df)
}

func TestDetermineRenderer(t *testing.T) {
	// TODO:  Currently this will ALWAYS return a DinghyfileParser; when we
	//        support additional types, we'll need to add those tests here.
	b := testPipelineBuilder()
	r := b.DetermineParser("dinghyfile")
	assert.Equal(t, "*dinghyfile.DinghyfileParser", reflect.TypeOf(r).String())
}

func TestGetPipelineByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	emptyset := []plank.Pipeline{}
	foundset := []plank.Pipeline{
		plank.Pipeline{Name: "pipelineName", ID: "pipelineID"},
	}

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetPipelines(gomock.Eq("testapp"),"").Return(emptyset, nil).Times(1)
	client.EXPECT().GetPipelines(gomock.Eq("testapp"), "").Return(foundset, nil).Times(1)
	client.EXPECT().UpsertPipeline(gomock.Any(), gomock.Eq(""), "").Times(1)

	b := testPipelineBuilder()
	b.Client = client

	res, err := b.GetPipelineByID("testapp", "pipelineName")
	assert.Nil(t, err)
	assert.Equal(t, "pipelineID", res)
}

func TestGetPipelineByIDUpsertFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	emptyset := []plank.Pipeline{}
	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetPipelines(gomock.Eq("testapp"), "").Return(emptyset, nil).Times(1)
	client.EXPECT().UpsertPipeline(gomock.Any(), gomock.Eq(""), "").Return(errors.New("upsert fail test")).Times(1)

	b := testPipelineBuilder()
	b.Client = client

	res, err := b.GetPipelineByID("testapp", "pipelineName")
	assert.NotNil(t, err)
	assert.Equal(t, "upsert fail test", err.Error())
	assert.Equal(t, "", res)
}

func TestUpdatePipelinesDeleteStaleWithExisting(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	existingPipeline := plank.Pipeline{Name: "ExistingPipeline", ID: "ExistingID", Application: "testapp", Locked: &plank.PipelineLockType{true, true}}
	deletedPipeline := plank.Pipeline{Name: "DeletedPipeline", ID: "DeletedID", Application: "testapp"}
	newPipeline := plank.Pipeline{Name: "NewPipeline", ID: "NewID", Locked: &plank.PipelineLockType{true, true}}

	existing := []plank.Pipeline{existingPipeline, deletedPipeline}
	newPipelines := []plank.Pipeline{existingPipeline, newPipeline}
	combined := []plank.Pipeline{existingPipeline, deletedPipeline, newPipeline}

	testapp := &plank.Application{Name: "testapp"}

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetApplication("testapp", "").Return(nil, nil).Times(1)
	client.EXPECT().GetPipelines(gomock.Eq("testapp"), "").Return(existing, nil).Times(1)
	client.EXPECT().GetPipelines(gomock.Eq("testapp"), "").Return(combined, nil).Times(1)
	client.EXPECT().UpsertPipeline(gomock.Eq(existingPipeline), gomock.Eq(existingPipeline.ID), "").Return(nil).Times(1)
	client.EXPECT().UpsertPipeline(gomock.Eq(newPipeline), gomock.Eq(newPipeline.ID), "").Return(nil).Times(1)
	client.EXPECT().DeletePipeline(gomock.Eq(deletedPipeline), "").Return(errors.New("fake delete failure")).Times(1)

	b := testPipelineBuilder()
	b.Client = client

	err := b.updatePipelines(testapp, newPipelines, true, "true")
	assert.Nil(t, err)
}

func TestUpdatePipelinesNoDeleteStaleWithExisting(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	existingPipeline := plank.Pipeline{Name: "ExistingPipeline", ID: "ExistingID", Application: "testapp", Locked: &plank.PipelineLockType{true, true}}
	deletedPipeline := plank.Pipeline{Name: "DeletedPipeline", ID: "DeletedID", Application: "testapp"}
	newPipeline := plank.Pipeline{Name: "NewPipeline", ID: "NewID", Locked: &plank.PipelineLockType{true, true}}

	existing := []plank.Pipeline{existingPipeline, deletedPipeline}
	newPipelines := []plank.Pipeline{existingPipeline, newPipeline}

	testapp := &plank.Application{Name: "testapp"}

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetApplication("testapp", "").Return(nil, nil).Times(1)
	client.EXPECT().GetPipelines(gomock.Eq("testapp"), "").Return(existing, nil).Times(1)
	client.EXPECT().UpsertPipeline(gomock.Eq(existingPipeline), gomock.Eq(existingPipeline.ID), "").Return(nil).Times(1)
	client.EXPECT().UpsertPipeline(gomock.Eq(newPipeline), gomock.Eq(newPipeline.ID), "").Return(nil).Times(1)
	client.EXPECT().DeletePipeline(gomock.Eq(deletedPipeline), "").Return(errors.New("fake delete failure")).Times(0)

	b := testPipelineBuilder()
	b.Client = client

	err := b.updatePipelines(testapp, newPipelines, false, "true")
	assert.Nil(t, err)
}

func TestUpdatePipelinesDeleteStaleWithFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	existingPipeline := plank.Pipeline{Name: "ExistingPipeline", ID: "ExistingID", Application: "testapp", Locked: &plank.PipelineLockType{true, true}}
	deletedPipeline := plank.Pipeline{Name: "DeletedPipeline", ID: "DeletedID", Application: "testapp"}
	newPipeline := plank.Pipeline{Name: "NewPipeline", ID: "NewID", Locked: &plank.PipelineLockType{true, true}}

	existing := []plank.Pipeline{existingPipeline, deletedPipeline}
	newPipelines := []plank.Pipeline{existingPipeline, newPipeline}

	testapp := &plank.Application{Name: "testapp"}

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetApplication("testapp", "").Return(nil, nil).Times(1)
	client.EXPECT().GetPipelines(gomock.Eq("testapp"), "").Return(existing, nil).Times(1)
	client.EXPECT().GetPipelines(gomock.Eq("testapp"), "").Return(nil, errors.New("failure to retrieve")).Times(1)
	client.EXPECT().UpsertPipeline(gomock.Eq(existingPipeline), gomock.Eq(existingPipeline.ID), "").Return(nil).Times(1)
	client.EXPECT().UpsertPipeline(gomock.Eq(newPipeline), gomock.Eq(newPipeline.ID), "").Return(nil).Times(1)
	// Should not be called because we couldn't re-retrieve the combined list.
	client.EXPECT().DeletePipeline(gomock.Eq(deletedPipeline), "").Return(errors.New("fake delete failure")).Times(0)

	b := testPipelineBuilder()
	b.Client = client

	err := b.updatePipelines(testapp, newPipelines, true, "true")
	assert.Nil(t, err)
}

func TestUpdatePipelinesUpsertFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	existingPipeline := plank.Pipeline{Name: "ExistingPipeline", ID: "ExistingID", Application: "testapp", Locked: &plank.PipelineLockType{true, true}}
	deletedPipeline := plank.Pipeline{Name: "DeletedPipeline", ID: "DeletedID", Application: "testapp"}
	newPipeline := plank.Pipeline{Name: "NewPipeline", ID: "NewID", Locked: &plank.PipelineLockType{true, true}}

	existing := []plank.Pipeline{existingPipeline, deletedPipeline}
	newPipelines := []plank.Pipeline{existingPipeline, newPipeline}

	testapp := &plank.Application{Name: "testapp"}

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetApplication("testapp", "").Return(nil, nil).Times(1)
	client.EXPECT().GetPipelines(gomock.Eq("testapp"), "").Return(existing, nil).Times(1)
	client.EXPECT().UpsertPipeline(gomock.Eq(existingPipeline), gomock.Eq(existingPipeline.ID), "").Return(nil).Times(1)
	client.EXPECT().UpsertPipeline(gomock.Eq(newPipeline), gomock.Eq(newPipeline.ID), "").Return(errors.New("upsert fail test")).Times(1)
	// Should not get called at all, because of earlier error
	client.EXPECT().DeletePipeline(gomock.Eq(deletedPipeline), "").Times(0)

	b := testPipelineBuilder()
	b.Client = client

	err := b.updatePipelines(testapp, newPipelines, true, "true")
	assert.NotNil(t, err)
	assert.Equal(t, "upsert fail test", err.Error())
}

func TestUpdatePipelinesRespectsAutoLockOn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// New pipeline from file has locks "false"
	newPipeline := plank.Pipeline{Name: "NewPipeline", ID: "NewID", Locked: &plank.PipelineLockType{false, false}}
	expectedPipeline := plank.Pipeline{}
	copier.Copy(&expectedPipeline, &newPipeline)
	// Expect to upsert with locks "true"
	expectedPipeline.Locked = &plank.PipelineLockType{true, true}

	testapp := &plank.Application{Name: "testapp"}

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetApplication("testapp","").Return(nil, nil).Times(1)
	client.EXPECT().GetPipelines(gomock.Eq("testapp"), "").Return([]plank.Pipeline{}, nil).Times(1)
	client.EXPECT().UpsertPipeline(gomock.Eq(expectedPipeline), gomock.Eq(newPipeline.ID), "").Return(nil).Times(1)

	b := testPipelineBuilder()
	b.Client = client

	err := b.updatePipelines(testapp, []plank.Pipeline{newPipeline}, false, "true")
	assert.Nil(t, err)
}

func TestUpdatePipelinesRespectsAutoLockOff(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// New pipeline from file has locks "false"
	newPipeline := plank.Pipeline{Name: "NewPipeline", ID: "NewID", Locked: &plank.PipelineLockType{false, false}}
	expectedPipeline := plank.Pipeline{}
	copier.Copy(&expectedPipeline, &newPipeline)
	// Expect to upsert with locks "true"
	expectedPipeline.Locked = &plank.PipelineLockType{false, false}

	testapp := &plank.Application{Name: "testapp"}

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetApplication("testapp", "").Return(nil, nil).Times(1)
	client.EXPECT().GetPipelines(gomock.Eq("testapp"), "").Return([]plank.Pipeline{}, nil).Times(1)
	client.EXPECT().UpsertPipeline(gomock.Eq(expectedPipeline), gomock.Eq(newPipeline.ID), "").Return(nil).Times(1)

	b := testPipelineBuilder()
	b.Client = client

	err := b.updatePipelines(testapp, []plank.Pipeline{newPipeline}, false, "")
	assert.Nil(t, err)
}

func TestRebuildModuleRoots(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jsonOne := `{ "application": "testone"}`

	roots := []string{
		"https://github.com/repos/org/repo/contents/dinghyfile?ref=branch",
	}

	b := testPipelineBuilder()
	url := b.Downloader.EncodeURL("org", "repo", "template_repo", "branch")

	depman := NewMockDependencyManager(ctrl)
	depman.EXPECT().GetRoots(gomock.Eq(url)).Return(roots).Times(1)
	b.Depman = depman

	renderer := NewMockParser(ctrl)
	renderer.EXPECT().Parse(gomock.Eq("org"), gomock.Eq("repo"), gomock.Eq("dinghyfile"), gomock.Eq("branch"), gomock.Nil()).Return(bytes.NewBufferString(jsonOne), nil).Times(1)
	b.Parser = renderer

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetApplication(gomock.Eq("testone"), "").Return(nil, nil).Times(1)
	client.EXPECT().GetPipelines(gomock.Eq("testone"), "").Return([]plank.Pipeline{}, nil).Times(1)
	b.Client = client
	b.DinghyfileName = "dinghyfile"
	b.RepositoryRawdataProcessing = false

	err := b.RebuildModuleRoots("org", "repo", "template_repo", "branch")
	assert.Nil(t, err)
}

func TestRebuildModuleRootsProcessTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jsonOne := `{
    "application": "testone"
    }`

	rawdata := `{
	"testkey": "testval"
	}`

	roots := []string{
		"https://github.com/repos/org/repo/contents/dinghyfile?ref=branch",
	}

	b := testPipelineBuilder()
	url := b.Downloader.EncodeURL("org", "repo", "template_repo", "branch")

	depman := NewMockDependencyManager(ctrl)
	depman.EXPECT().GetRoots(gomock.Eq(url)).Return(roots).Times(1)
	depman.EXPECT().GetRawData(gomock.Eq(roots[0])).Return(rawdata, nil).Times(1)
	b.Depman = depman

	renderer := NewMockParser(ctrl)
	renderer.EXPECT().Parse(gomock.Eq("org"), gomock.Eq("repo"), gomock.Eq("dinghyfile"), gomock.Eq("branch"), gomock.Nil()).Return(bytes.NewBufferString(jsonOne), nil).Times(1)
	b.Parser = renderer

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetApplication(gomock.Eq("testone"), "").Return(nil, nil).Times(1)
	client.EXPECT().GetPipelines(gomock.Eq("testone"), "").Return([]plank.Pipeline{}, nil).Times(1)
	b.Client = client
	b.DinghyfileName = "dinghyfile"
	b.RepositoryRawdataProcessing = true

	err := b.RebuildModuleRoots("org", "repo", "template_repo", "branch")
	assert.Nil(t, err)
}

func TestRebuildModuleRootsFailureCase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jsonTwo := `{ "application": "testtwo"}`

	roots := []string{
		"https://github.com/repos/org/repo/contents/dinghyfile?ref=branch",
		"https://github.com/repos/org/repo2/contents/dinghyfile?ref=branch",
	}

	b := testPipelineBuilder()
	url := b.Downloader.EncodeURL("org", "repo", "template_repo", "branch")

	depman := NewMockDependencyManager(ctrl)
	depman.EXPECT().GetRoots(gomock.Eq(url)).Return(roots).Times(1)
	b.Depman = depman

	renderer := NewMockParser(ctrl)
	renderer.EXPECT().Parse(gomock.Eq("org"), gomock.Eq("repo"), gomock.Eq("dinghyfile"), gomock.Eq("branch"), gomock.Nil()).Return(nil, errors.New("rebuild fail test")).Times(1)
	renderer.EXPECT().Parse(gomock.Eq("org"), gomock.Eq("repo2"), gomock.Eq("dinghyfile"), gomock.Eq("branch"), gomock.Nil()).Return(bytes.NewBufferString(jsonTwo), nil).Times(1)
	b.Parser = renderer

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetApplication(gomock.Eq("testtwo"), "").Return(nil, nil).Times(1)
	client.EXPECT().GetPipelines(gomock.Eq("testtwo"), "").Return([]plank.Pipeline{}, nil).Times(1)
	b.Client = client
	b.DinghyfileName = "dinghyfile"
	b.RepositoryRawdataProcessing = false

	err := b.RebuildModuleRoots("org", "repo", "template_repo", "branch")
	assert.NotNil(t, err)
	assert.Equal(t, "Not all upstream dinghyfiles were updated successfully", err.Error())
}

func TestAddRenderer(t *testing.T) {
	b := testPipelineBuilder()
	b.AddUnmarshaller(&DinghyJsonUnmarshaller{})
	assert.Equal(t, len(b.Ums), 2)
}

type mockNotifier struct {
	SuccessCalls int
	FailureCalls int
	LastError    error
}

func (m *mockNotifier) SendSuccess(org, repo, path string, notificationsType plank.NotificationsType, content map[string]interface{}) {
	m.SuccessCalls = m.SuccessCalls + 1
}
func (m *mockNotifier) SendFailure(org, repo, path string, err error, notificationsType plank.NotificationsType, content map[string]interface{}) {
	m.FailureCalls = m.FailureCalls + 1
	m.LastError = err
}

func (m *mockNotifier) SendOnValidation() bool {
	return true
}

func TestSuccessNotifier(t *testing.T) {
	b := testPipelineBuilder()
	n := mockNotifier{}
	b.Notifiers = []notifiers.Notifier{&n}
	b.NotifySuccess("foo", "bar", "biff", nil)
	assert.Equal(t, n.SuccessCalls, 1)
	assert.Equal(t, n.FailureCalls, 0)
}

func TestFailureNotifier(t *testing.T) {
	b := testPipelineBuilder()
	n := mockNotifier{}
	b.Notifiers = []notifiers.Notifier{&n}
	b.NotifyFailure("foo", "bar", "biff", errors.New("foo"), "")
	assert.Equal(t, n.SuccessCalls, 0)
	assert.Equal(t, n.FailureCalls, 1)
	assert.Equal(t, n.LastError.Error(), "foo")
}

func Test_extractApplicationName(t *testing.T) {
	tests := []struct {
		name       string
		dinghyfile string
		want       string
		wantErr    bool
	}{
		{
			name:       "json_test",
			dinghyfile: `{"application": "test_app_name"}`,
			want:       "test_app_name",
			wantErr:    false,
		},
		{
			name: "yaml_test",
			dinghyfile: `application: "my-awesome-application"
# You can do inline comments now
globals:
  waitTime: 42
  retries: 5
pipelines:
- application: "my-awesome-application"
  name: "My cool pipeline"
  appConfig: {}
  keepWaitingPipelines: false
  limitConcurrent: true
  stages:
  - name: Wait For It...
    refId: '1'
    requisiteStageRefIds: []
    type: wait
    waitTime: 4
  {{ module "some.stage.module" "something" }}
  triggers: []`,
			want:    "my-awesome-application",
			wantErr: false,
		},
		{
			name: "hcl_test",
			dinghyfile: `"application" = "some-app"
"globals" = {
    "waitTime" = 42
}`,
			want:    "some-app",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractApplicationName(tt.dinghyfile)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractApplicationName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractApplicationName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPipelineBuilder_getContent(t *testing.T) {
	type fields struct {
		Downloader                  Downloader
		Depman                      DependencyManager
		TemplateRepo                string
		TemplateOrg                 string
		DinghyfileName              string
		Client                      util.PlankClient
		DeleteStalePipelines        bool
		AutolockPipelines           string
		EventClient                 events.EventClient
		Parser                      Parser
		Logger                      log.DinghyLog
		Ums                         []Unmarshaller
		Notifiers                   []notifiers.Notifier
		PushRaw                     map[string]interface{}
		GlobalVariablesMap          map[string]interface{}
		RepositoryRawdataProcessing bool
		RebuildingModules           bool
		Action                      pipebuilder.BuilderAction
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]interface{}
	}{
		{
			name: "Content should return a map populated with 'raw' property if no Log Events are available",
			fields: fields{
				Logger: NewDinghylog(),
				PushRaw: map[string]interface{}{
					"head_commit": map[string]interface{}{
						"id": "a5fc63bd5a8bdb342d1e83933a5b5c99010e61e4",
					},
				},
			},
			want: map[string]interface{}{
				"rawdata": map[string]interface{}{
					"head_commit": map[string]interface{}{
						"id": "a5fc63bd5a8bdb342d1e83933a5b5c99010e61e4",
					},
				},
			},
		},
		{
			name: "Content should return a map populated with 'raw' and 'logevent' properties, since both are populated.",
			fields: fields{
				Logger: func() log.DinghyLog {
					return NewDinghylogWithContent("test")
				}(),
				PushRaw: map[string]interface{}{},
			},
			want: map[string]interface{}{
				"rawdata":  map[string]interface{}{},
				"logevent": "test",
			},
		},
		{
			name: "Content should return a map populated with 'logevent' properties, since no raw data is.",
			fields: fields{
				Logger: func() log.DinghyLog {
					return NewDinghylogWithContent("test")
				}(),
				PushRaw: nil,
			},
			want: map[string]interface{}{
				"logevent": "test",
			},
		},
		{
			name: "Content should return a empty map, since no properties are populated.",
			fields: fields{
				Logger: NewDinghylog(),
			},
			want: map[string]interface{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &PipelineBuilder{
				Downloader:                  tt.fields.Downloader,
				Depman:                      tt.fields.Depman,
				TemplateRepo:                tt.fields.TemplateRepo,
				TemplateOrg:                 tt.fields.TemplateOrg,
				DinghyfileName:              tt.fields.DinghyfileName,
				Client:                      tt.fields.Client,
				DeleteStalePipelines:        tt.fields.DeleteStalePipelines,
				AutolockPipelines:           tt.fields.AutolockPipelines,
				EventClient:                 tt.fields.EventClient,
				Parser:                      tt.fields.Parser,
				Logger:                      tt.fields.Logger,
				Ums:                         tt.fields.Ums,
				Notifiers:                   tt.fields.Notifiers,
				PushRaw:                     tt.fields.PushRaw,
				GlobalVariablesMap:          tt.fields.GlobalVariablesMap,
				RepositoryRawdataProcessing: tt.fields.RepositoryRawdataProcessing,
				RebuildingModules:           tt.fields.RebuildingModules,
				Action:                      tt.fields.Action,
			}
			if got := b.getNotificationContent(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getNotificationContent() = %v, want %v", got, tt.want)
			}
		})
	}
}
