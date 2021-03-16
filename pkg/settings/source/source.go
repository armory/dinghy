package source

import "github.com/armory/dinghy/pkg/settings/lighthouse"

//go:generate stringer -type=SettingField
type SettingField int

const (
	TemplateOrg SettingField = iota
	TemplateRepo
	DinghyFilename
	AutoLockPipelines
	SpinnakerUIURL
	GitHubCredsPath
	GitHubToken
	GithubEndpoint
	GitLabToken
	GitLabEndpoint
	StashCredsPath
	StashUsername
	StashToken
	StashEndpoint
	FiatUser
	Logging
	Secrets
	ParserFormat
	RepoConfig
	Orca
	Front50
	Deck
	Echo
	Fiat
	Redis
	spinnakerSupplied
	Server
	Http
	WebhookValidations
	WebhookValidationEnabledProviders
	RepositoryRawdataProcessing
	LogEventTTLMinutes
	SQL
)

type Source interface {
	Load() (*lighthouse.Settings, error)
	GetConfigurationByKey(SettingField) interface{}
	GetSourceName() string
}
