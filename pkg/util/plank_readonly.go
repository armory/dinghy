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

import "github.com/armory/plank/v3"

type PlankReadOnly struct {
	Plank *plank.Client
}

func (p PlankReadOnly) GetApplication(string string) (*plank.Application, error){
	return p.Plank.GetApplication(string)
}

func (p PlankReadOnly) UpdateApplicationNotifications(plank.NotificationsType, string) error{
	return nil
}

func (p PlankReadOnly) GetApplicationNotifications(app string) (*plank.NotificationsType, error){
	return p.Plank.GetApplicationNotifications(app)
}

func (p PlankReadOnly) CreateApplication(*plank.Application) error{
	return nil
}

func (p PlankReadOnly) UpdateApplication(plank.Application) error{
	return nil
}

func (p PlankReadOnly) GetPipelines(string string) ([]plank.Pipeline, error){
	return p.Plank.GetPipelines(string)
}

func (p PlankReadOnly) DeletePipeline(plank.Pipeline) error{
	return nil
}

func (p PlankReadOnly) UpsertPipeline(pipe plank.Pipeline, str string) error{
	return nil
}

func (p PlankReadOnly) ResyncFiat() error{
	return nil
}

func (p PlankReadOnly) ArmoryEndpointsEnabled() bool{
	return false
}

func (p PlankReadOnly) EnableArmoryEndpoints(){
}
