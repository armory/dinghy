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
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/assert"

	"github.com/armory/plank"

	"github.com/armory/dinghy/pkg/mock"
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

	logger := mock.NewMockFieldLogger(ctrl)
	logger.EXPECT().Infof(gomock.Eq("Unmarshalled: %v"), gomock.Any()).Times(1)
	logger.EXPECT().Infof(gomock.Eq("Found pipelines for %v: %v"), gomock.Any()).Times(1)
	logger.EXPECT().Infof(gomock.Eq("Dinghyfile struct: %v"), gomock.Any()).Times(1)
	logger.EXPECT().Infof(gomock.Eq("Updated: %s"), gomock.Any()).Times(1)
	logger.EXPECT().Infof(gomock.Eq("Rendered: %s"), gomock.Any()).Times(1)
	logger.EXPECT().Info(gomock.Eq("Looking up existing pipelines")).Times(1)

	// Because we've set the renderer, we should NOT get this message...
	logger.EXPECT().Info(gomock.Eq("Calling DetermineRenderer")).Times(0)

	// Never gets UpsertPipeline because our Render returns no pipelines.
	client.EXPECT().UpsertPipeline(gomock.Any(), "").Return(nil).Times(0)
	pb := testPipelineBuilder()
	pb.Renderer = renderer
	pb.Client = client
	pb.Logger = logger
	assert.Nil(t, pb.ProcessDinghyfile("myorg", "myrepo", "the/full/path"))
}

// Note: This ALSO tests the error case where the renderer fails (in this
// example, because the rendered file path info is invalid)
func TestProcessDinghyfileDefaultRenderer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockFieldLogger(ctrl)
	logger.EXPECT().Info("Calling DetermineRenderer").Times(1)
	logger.EXPECT().Error(gomock.Eq("Failed to download")).Times(1)
	logger.EXPECT().Errorf(gomock.Eq("Failed to render dinghyfile %s: %s"), gomock.Eq("notfound"), gomock.Eq("File not found")).Times(1)

	pb := testPipelineBuilder()
	pb.Logger = logger
	res := pb.ProcessDinghyfile("fake", "news", "notfound")
	assert.NotNil(t, res)
}

func TestProcessDinghyfileFailedUnmarshal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rendered := `{blargh}`

	renderer := NewMockRenderer(ctrl)
	renderer.EXPECT().Render(gomock.Eq("myorg"), gomock.Eq("myrepo"), gomock.Eq("the/full/path"), gomock.Any()).Return(bytes.NewBuffer([]byte(rendered)), nil).Times(1)

	logger := mock.NewMockFieldLogger(ctrl)
	logger.EXPECT().Errorf(gomock.Eq("UpdateDinghyfile malformed json: %s"), gomock.Any()).Times(1)
	logger.EXPECT().Errorf(gomock.Eq("Failed to update dinghyfile %s: %s"), gomock.Any()).Times(1)
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	pb := testPipelineBuilder()
	pb.Logger = logger
	pb.Renderer = renderer
	res := pb.ProcessDinghyfile("myorg", "myrepo", "the/full/path")
	assert.NotNil(t, res)
}

func TestProcessDinghyfileFailedUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rendered := `{"application": "testapp"}`

	renderer := NewMockRenderer(ctrl)
	renderer.EXPECT().Render(gomock.Eq("myorg"), gomock.Eq("myrepo"), gomock.Eq("the/full/path"), gomock.Any()).Return(bytes.NewBuffer([]byte(rendered)), nil).Times(1)

	logger := mock.NewMockFieldLogger(ctrl)
	logger.EXPECT().Infof(gomock.Eq("Creating application '%s'..."), gomock.Eq("testapp")).Times(1)
	logger.EXPECT().Errorf("Failed to create application (%s)", gomock.Any())
	logger.EXPECT().Errorf(gomock.Eq("Failed to update Pipelines for %s: %s"), gomock.Eq("the/full/path")).Times(1)
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetApplication(gomock.Eq("testapp")).Return(nil, errors.New("not found")).Times(1)
	client.EXPECT().CreateApplication(gomock.Any()).Return(errors.New("boom")).Times(1)

	pb := testPipelineBuilder()
	pb.Logger = logger
	pb.Renderer = renderer
	pb.Client = client
	res := pb.ProcessDinghyfile("myorg", "myrepo", "the/full/path")
	assert.NotNil(t, res)
	assert.Equal(t, "boom", res.Error())
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
			d, _ := b.UpdateDinghyfile(c.dinghyRaw)
			assert.Equal(t, d, c.dinghyStruct)
		})
	}
}

func TestUpdateDinghyfileMalformed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	b := testPipelineBuilder()
	logger := mock.NewMockFieldLogger(ctrl)
	b.Logger = logger
	logger.EXPECT().Errorf(gomock.Eq("UpdateDinghyfile malformed json: %s"), gomock.Any()).Times(1)

	df, err := b.UpdateDinghyfile([]byte("{ garbage"))
	assert.Equal(t, ErrMalformedJSON, err)
	assert.NotNil(t, df)
}

func TestDetermineRenderer(t *testing.T) {
	// TODO:  Currently this will ALWAYS return a DinghyfileRenderer; when we
	//        support additional types, we'll need to add those tests here.
	b := testPipelineBuilder()
	r := b.DetermineRenderer("dinghyfile")
	assert.Equal(t, "*dinghyfile.DinghyfileRenderer", reflect.TypeOf(r).String())
}

func TestGetPipelineByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	emptyset := []plank.Pipeline{}
	foundset := []plank.Pipeline{
		plank.Pipeline{Name: "pipelineName", ID: "pipelineID"},
	}

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetPipelines(gomock.Eq("testapp")).Return(emptyset, nil).Times(1)
	client.EXPECT().GetPipelines(gomock.Eq("testapp")).Return(foundset, nil).Times(1)
	client.EXPECT().UpsertPipeline(gomock.Any(), gomock.Eq("")).Times(1)

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
	client.EXPECT().GetPipelines(gomock.Eq("testapp")).Return(emptyset, nil).Times(1)
	client.EXPECT().UpsertPipeline(gomock.Any(), gomock.Eq("")).Return(errors.New("upsert fail test")).Times(1)

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

	existingPipeline := plank.Pipeline{Name: "ExistingPipeline", ID: "ExistingID", Application: "testapp", Locked: plank.PipelineLockType{true, true}}
	deletedPipeline := plank.Pipeline{Name: "DeletedPipeline", ID: "DeletedID", Application: "testapp"}
	newPipeline := plank.Pipeline{Name: "NewPipeline", ID: "NewID", Locked: plank.PipelineLockType{true, true}}

	existing := []plank.Pipeline{existingPipeline, deletedPipeline}
	newPipelines := []plank.Pipeline{existingPipeline, newPipeline}
	combined := []plank.Pipeline{existingPipeline, deletedPipeline, newPipeline}

	testapp := &plank.Application{Name: "testapp"}

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetApplication("testapp").Return(nil, nil).Times(1)
	client.EXPECT().GetPipelines(gomock.Eq("testapp")).Return(existing, nil).Times(1)
	client.EXPECT().GetPipelines(gomock.Eq("testapp")).Return(combined, nil).Times(1)
	client.EXPECT().UpsertPipeline(gomock.Eq(existingPipeline), gomock.Eq(existingPipeline.ID)).Return(nil).Times(1)
	client.EXPECT().UpsertPipeline(gomock.Eq(newPipeline), gomock.Eq(newPipeline.ID)).Return(nil).Times(1)
	client.EXPECT().DeletePipeline(gomock.Eq(deletedPipeline)).Return(errors.New("fake delete failure")).Times(1)

	b := testPipelineBuilder()
	b.Client = client

	err := b.updatePipelines(testapp, newPipelines, true, "true")
	assert.Nil(t, err)
}

func TestUpdatePipelinesNoDeleteStaleWithExisting(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	existingPipeline := plank.Pipeline{Name: "ExistingPipeline", ID: "ExistingID", Application: "testapp", Locked: plank.PipelineLockType{true, true}}
	deletedPipeline := plank.Pipeline{Name: "DeletedPipeline", ID: "DeletedID", Application: "testapp"}
	newPipeline := plank.Pipeline{Name: "NewPipeline", ID: "NewID", Locked: plank.PipelineLockType{true, true}}

	existing := []plank.Pipeline{existingPipeline, deletedPipeline}
	newPipelines := []plank.Pipeline{existingPipeline, newPipeline}

	testapp := &plank.Application{Name: "testapp"}

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetApplication("testapp").Return(nil, nil).Times(1)
	client.EXPECT().GetPipelines(gomock.Eq("testapp")).Return(existing, nil).Times(1)
	client.EXPECT().UpsertPipeline(gomock.Eq(existingPipeline), gomock.Eq(existingPipeline.ID)).Return(nil).Times(1)
	client.EXPECT().UpsertPipeline(gomock.Eq(newPipeline), gomock.Eq(newPipeline.ID)).Return(nil).Times(1)
	client.EXPECT().DeletePipeline(gomock.Eq(deletedPipeline)).Return(errors.New("fake delete failure")).Times(0)

	b := testPipelineBuilder()
	b.Client = client

	err := b.updatePipelines(testapp, newPipelines, false, "true")
	assert.Nil(t, err)
}

func TestUpdatePipelinesDeleteStaleWithFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	existingPipeline := plank.Pipeline{Name: "ExistingPipeline", ID: "ExistingID", Application: "testapp", Locked: plank.PipelineLockType{true, true}}
	deletedPipeline := plank.Pipeline{Name: "DeletedPipeline", ID: "DeletedID", Application: "testapp"}
	newPipeline := plank.Pipeline{Name: "NewPipeline", ID: "NewID", Locked: plank.PipelineLockType{true, true}}

	existing := []plank.Pipeline{existingPipeline, deletedPipeline}
	newPipelines := []plank.Pipeline{existingPipeline, newPipeline}

	testapp := &plank.Application{Name: "testapp"}

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetApplication("testapp").Return(nil, nil).Times(1)
	client.EXPECT().GetPipelines(gomock.Eq("testapp")).Return(existing, nil).Times(1)
	client.EXPECT().GetPipelines(gomock.Eq("testapp")).Return(nil, errors.New("failure to retrieve")).Times(1)
	client.EXPECT().UpsertPipeline(gomock.Eq(existingPipeline), gomock.Eq(existingPipeline.ID)).Return(nil).Times(1)
	client.EXPECT().UpsertPipeline(gomock.Eq(newPipeline), gomock.Eq(newPipeline.ID)).Return(nil).Times(1)
	// Should not be called because we couldn't re-retrieve the combined list.
	client.EXPECT().DeletePipeline(gomock.Eq(deletedPipeline)).Return(errors.New("fake delete failure")).Times(0)

	b := testPipelineBuilder()
	b.Client = client

	err := b.updatePipelines(testapp, newPipelines, true, "true")
	assert.Nil(t, err)
}

func TestUpdatePipelinesUpsertFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	existingPipeline := plank.Pipeline{Name: "ExistingPipeline", ID: "ExistingID", Application: "testapp", Locked: plank.PipelineLockType{true, true}}
	deletedPipeline := plank.Pipeline{Name: "DeletedPipeline", ID: "DeletedID", Application: "testapp"}
	newPipeline := plank.Pipeline{Name: "NewPipeline", ID: "NewID", Locked: plank.PipelineLockType{true, true}}

	existing := []plank.Pipeline{existingPipeline, deletedPipeline}
	newPipelines := []plank.Pipeline{existingPipeline, newPipeline}

	testapp := &plank.Application{Name: "testapp"}

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetApplication("testapp").Return(nil, nil).Times(1)
	client.EXPECT().GetPipelines(gomock.Eq("testapp")).Return(existing, nil).Times(1)
	client.EXPECT().UpsertPipeline(gomock.Eq(existingPipeline), gomock.Eq(existingPipeline.ID)).Return(nil).Times(1)
	client.EXPECT().UpsertPipeline(gomock.Eq(newPipeline), gomock.Eq(newPipeline.ID)).Return(errors.New("upsert fail test")).Times(1)
	// Should not get called at all, because of earlier error
	client.EXPECT().DeletePipeline(gomock.Eq(deletedPipeline)).Times(0)

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
	newPipeline := plank.Pipeline{Name: "NewPipeline", ID: "NewID", Locked: plank.PipelineLockType{false, false}}
	expectedPipeline := plank.Pipeline{}
	copier.Copy(&expectedPipeline, &newPipeline)
	// Expect to upsert with locks "true"
	expectedPipeline.Locked = plank.PipelineLockType{true, true}

	testapp := &plank.Application{Name: "testapp"}

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetApplication("testapp").Return(nil, nil).Times(1)
	client.EXPECT().GetPipelines(gomock.Eq("testapp")).Return([]plank.Pipeline{}, nil).Times(1)
	client.EXPECT().UpsertPipeline(gomock.Eq(expectedPipeline), gomock.Eq(newPipeline.ID)).Return(nil).Times(1)

	b := testPipelineBuilder()
	b.Client = client

	err := b.updatePipelines(testapp, []plank.Pipeline{newPipeline}, false, "true")
	assert.Nil(t, err)
}

func TestUpdatePipelinesRespectsAutoLockOff(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// New pipeline from file has locks "false"
	newPipeline := plank.Pipeline{Name: "NewPipeline", ID: "NewID", Locked: plank.PipelineLockType{false, false}}
	expectedPipeline := plank.Pipeline{}
	copier.Copy(&expectedPipeline, &newPipeline)
	// Expect to upsert with locks "true"
	expectedPipeline.Locked = plank.PipelineLockType{false, false}

	testapp := &plank.Application{Name: "testapp"}

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetApplication("testapp").Return(nil, nil).Times(1)
	client.EXPECT().GetPipelines(gomock.Eq("testapp")).Return([]plank.Pipeline{}, nil).Times(1)
	client.EXPECT().UpsertPipeline(gomock.Eq(expectedPipeline), gomock.Eq(newPipeline.ID)).Return(nil).Times(1)

	b := testPipelineBuilder()
	b.Client = client

	err := b.updatePipelines(testapp, []plank.Pipeline{newPipeline}, false, "")
	assert.Nil(t, err)
}

func TestRebuildModuleRoots(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jsonOne := `{ "application": "testone" }`

	roots := []string{
		"https://github.com/repos/org/repo/contents/foo",
	}

	b := testPipelineBuilder()
	url := b.Downloader.EncodeURL("org", "repo", "rebuild_test")

	depman := NewMockDependencyManager(ctrl)
	depman.EXPECT().GetRoots(gomock.Eq(url)).Return(roots).Times(1)
	b.Depman = depman

	renderer := NewMockRenderer(ctrl)
	renderer.EXPECT().Render(gomock.Eq("org"), gomock.Eq("repo"), gomock.Eq("foo"), gomock.Nil()).Return(bytes.NewBufferString(jsonOne), nil).Times(1)
	b.Renderer = renderer

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetApplication(gomock.Eq("testone")).Return(nil, nil).Times(1)
	client.EXPECT().GetPipelines(gomock.Eq("testone")).Return([]plank.Pipeline{}, nil).Times(1)
	b.Client = client

	err := b.RebuildModuleRoots("org", "repo", "rebuild_test")
	assert.Nil(t, err)
}

func TestRebuildModuleRootsFailureCase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jsonTwo := `{ "application": "testtwo" }`

	roots := []string{
		"https://github.com/repos/org/repo/contents/foo",
		"https://github.com/repos/org/repo/contents/bar",
	}

	b := testPipelineBuilder()
	url := b.Downloader.EncodeURL("org", "repo", "rebuild_test")

	depman := NewMockDependencyManager(ctrl)
	depman.EXPECT().GetRoots(gomock.Eq(url)).Return(roots).Times(1)
	b.Depman = depman

	renderer := NewMockRenderer(ctrl)
	renderer.EXPECT().Render(gomock.Eq("org"), gomock.Eq("repo"), gomock.Eq("foo"), gomock.Nil()).Return(nil, errors.New("rebuild fail test")).Times(1)
	renderer.EXPECT().Render(gomock.Eq("org"), gomock.Eq("repo"), gomock.Eq("bar"), gomock.Nil()).Return(bytes.NewBufferString(jsonTwo), nil).Times(1)
	b.Renderer = renderer

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetApplication(gomock.Eq("testtwo")).Return(nil, nil).Times(1)
	client.EXPECT().GetPipelines(gomock.Eq("testtwo")).Return([]plank.Pipeline{}, nil).Times(1)
	b.Client = client

	err := b.RebuildModuleRoots("org", "repo", "rebuild_test")
	assert.NotNil(t, err)
	assert.Equal(t, "Not all upstream dinghyfiles were updated successfully", err.Error())
}
