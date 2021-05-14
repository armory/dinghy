/*
* Copyright 2020 Armory, Inc.

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

package util

import (
	"fmt"
	"github.com/armory/plank/v4"
	"github.com/google/uuid"
)

type PlankReadOnly struct {
	Plank 		*plank.Client
	tempPipes	*[]plank.Pipeline
}

func (p *PlankReadOnly) GetApplication(string, traceparent string) (*plank.Application, error) {
	return p.Plank.GetApplication(string, traceparent)
}

func (p *PlankReadOnly) UpdateApplicationNotifications(plank.NotificationsType, string, string) error {
	return nil
}

func (p *PlankReadOnly) GetApplicationNotifications(app, traceparent string) (*plank.NotificationsType, error) {
	return p.Plank.GetApplicationNotifications(app, traceparent)
}

func (p *PlankReadOnly) CreateApplication(*plank.Application, string) error {
	return nil
}

func (p *PlankReadOnly) UpdateApplication(plank.Application, string) error {
	return nil
}

func (p *PlankReadOnly) GetPipelines(appName, traceparent string) ([]plank.Pipeline, error) {
	pipes,err := p.Plank.GetPipelines(appName, traceparent)
	if err != nil {
		return  pipes,err
	}
	// Here we will get the previously created pipelines
	if p.tempPipes != nil {
		for _, val := range *p.tempPipes{
			pipes = append(pipes, val)
		}
	}

	return pipes, nil
}

func (p *PlankReadOnly) DeletePipeline(plank.Pipeline, string) error {
	return nil
}

func (p *PlankReadOnly) UpsertPipeline(pipe plank.Pipeline, appName string, traceparent string) error {
	// This is getting a little complex
	// When a pipeline does not exists dinghy create it so it can be referenced
	// Its a recursive call so it loops forever if this temp pipeline is not created
	if p.tempPipes == nil {
		p.tempPipes = &[]plank.Pipeline{}
	}
	// Auto generate a dummy id
	pipe.ID = fmt.Sprintf("auto-generated-dummy-id-%v", uuid.New().String())
	(*p.tempPipes) = append(*p.tempPipes, pipe)
	return nil
}

func (p *PlankReadOnly) ResyncFiat(string) error {
	return nil
}

func (p *PlankReadOnly) ArmoryEndpointsEnabled() bool {
	return false
}

func (p *PlankReadOnly) EnableArmoryEndpoints() {
}

func (p *PlankReadOnly) UseGateEndpoints() {
}

func (p *PlankReadOnly) UseServiceEndpoints() {
}