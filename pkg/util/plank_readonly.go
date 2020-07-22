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
