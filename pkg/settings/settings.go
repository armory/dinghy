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

// Package settings is a single place to put all of the application settings.
package settings

import (
	"github.com/armory/go-yaml-tools/pkg/secrets"
	"github.com/armory/go-yaml-tools/pkg/tls/client"
	"github.com/armory/go-yaml-tools/pkg/tls/server"
	"github.com/jinzhu/copier"
)

// Settings contains all information needed to startup and run the dinghy service
type Settings struct {
	TemplateOrg       string       `json:"templateOrg,omitempty" yaml:"templateOrg"`
	TemplateRepo      string       `json:"templateRepo,omitempty" yaml:"templateRepo"`
	DinghyFilename    string       `json:"dinghyFilename,omitempty" yaml:"dinghyFilename"`
	AutoLockPipelines string       `json:"autoLockPipelines,omitempty" yaml:"autoLockPipelines"`
	SpinnakerUIURL    string       `json:"spinUIUrl,omitempty" yaml:"spinUIUrl"`
	GitHubCredsPath   string       `json:"githubCredsPath,omitempty" yaml:"githubCredsPath"`
	GitHubToken       string       `json:"githubToken,omitempty" yaml:"githubToken"`
	GithubEndpoint    string       `json:"githubEndpoint,omitempty" yaml:"githubEndpoint"`
	GitLabToken       string       `json:"gitlabToken,omitempty" yaml:"gitlabToken"`
	GitLabEndpoint    string       `json:"gitlabEndpoint,omitempty" yaml:"gitlabEndpoint"`
	StashCredsPath    string       `json:"stashCredsPath,omitempty" yaml:"stashCredsPath"`
	StashUsername     string       `json:"stashUsername,omitempty" yaml:"stashUsername"`
	StashToken        string       `json:"stashToken,omitempty" yaml:"stashToken"`
	StashEndpoint     string       `json:"stashEndpoint,omitempty" yaml:"stashEndpoint"`
	FiatUser                          string       `json:"fiatUser,omitempty" yaml:"fiatUser"`
	Logging                           Logging      `json:"logging,omitempty" yaml:"logging"`
	Secrets                           Secrets      `json:"secrets,omitempty" yaml:"secrets"`
	ParserFormat                      string       `json:"parserFormat,omitempty" yaml:"parserFormat"`
	RepoConfig                        []RepoConfig `json:"repoConfig,omitempty" yaml:"repoConfig"`
	spinnakerSupplied                 `mapstructure:",squash"`
	Server                            server.ServerConfig `json:"server" yaml:"server"`
	Http                              client.Config       `json:"http" yaml:"http"`
	WebhookValidations                []WebhookValidation `json:"webhookValidations,omitempty" yaml:"webhookValidations"`
	WebhookValidationEnabledProviders []string            `json:"webhookValidationEnabledProviders,omitempty" yaml:"webhookValidationEnabledProviders"`
	RepositoryRawdataProcessing       bool                `json:"repositoryRawdataProcessing,omitempty" yaml:"repositoryRawdataProcessing"`
}

type WebhookValidation struct {
	Enabled 				bool	`json:"enabled,omitempty" yaml:"enabled"`
	VersionControlProvider 	string	`json:"versionControlProvider,omitempty" yaml:"versionControlProvider"`
	Organization 			string	`json:"organization,omitempty" yaml:"organization"`
	Repo	 				string	`json:"repo,omitempty" yaml:"repo"`
	Secret 					string	`json:"secret,omitempty" yaml:"secret"`
}



type spinnakerSupplied struct {
	Orca    spinnakerService `json:"orca,omitempty" yaml:"orca"`
	Front50 spinnakerService `json:"front50,omitempty" yaml:"front50"`
	Deck    spinnakerService `json:"deck,omitempty" yaml:"deck"`
	Echo    spinnakerService `json:"echo,omitempty" yaml:"echo"`
	Fiat    fiat             `json:"fiat,omitempty" yaml:"fiat"`
	Redis   Redis            `json:"redis,omitempty" yaml:"redis"`
}

type Redis struct {
	BaseURL    string `json:"baseUrl,omitempty" yaml:"baseUrl"`
	Password   string `json:"password,omitempty" yaml:"password"`
	Connection string `json:"connection,omitempty" yaml:"connection"`
}

type fiat struct {
	AuthUser         string `json:"authUser,omitempty" yaml:"authUser"`
	spinnakerService `mapstructure:",squash"`
}

type spinnakerService struct {
	Enabled string `json:"enabled,omitempty" yaml:"enabled"`
	BaseURL string `json:"baseUrl,omitempty" yaml:"baseUrl"`
}

type Logging struct {
	File   string        `json:"file,omitempty" yaml:"file"`
	Level  string        `json:"level,omitempty" yaml:"level"`
	Remote RemoteLogging `json:"remote" yaml:"remote"`
}

type RemoteLogging struct {
	Enabled    bool   `json:"enabled" yaml:"remote"`
	Endpoint   string `json:"endpoint" yaml:"endpoint"`
	CustomerID string `json:"customerId" yaml:"customerId"`
	Version    string `json:"version" yaml:"version"`
}

type Secrets struct {
	Vault secrets.VaultConfig `json:"vault" yaml:"vault"`
}

type RepoConfig struct {
	Provider string `json:"provider,omitempty" yaml:"provider"`
	Repo     string `json:"repo,omitempty" yaml:"repo"`
	Branch   string `json:"branch,omitempty" yaml:"branch"`
}

func (s *Settings) GetRepoConfig(provider, repo string) *RepoConfig {
	for _, c := range s.RepoConfig {
		if c.Provider == provider && c.Repo == repo {
			return &c
		}
	}
	return nil
}

// Redacted returns a copy of the Settings object with all the sensitive
// fields **REDACTED**.
func (s *Settings) Redacted() *Settings {
	redacted := &Settings{}
	copier.Copy(&redacted, s)

	if redacted.GitHubToken != "" {
		redacted.GitHubToken = "**REDACTED**"
	}
	if redacted.GitLabToken != "" {
		redacted.GitLabToken = "**REDACTED**"
	}
	if redacted.StashToken != "" {
		redacted.StashToken = "**REDACTED**"
	}
	if redacted.Secrets.Vault.Token != "" {
		redacted.Secrets.Vault.Token = "**REDACTED**"
	}
	if redacted.spinnakerSupplied.Redis.Password != "" {
		redacted.spinnakerSupplied.Redis.Password = "**REDACTED**"
	}
	return redacted
}
