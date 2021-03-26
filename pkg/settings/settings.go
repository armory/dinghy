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
	"net/http"
	"time"
)

// Settings contains all information needed to startup and run the dinghy service
type Settings struct {
	// Organization account that will have the template repository
	TemplateOrg                       string       `json:"templateOrg,omitempty" yaml:"templateOrg"`
	// Repository for templates (modules)
	TemplateRepo                      string       `json:"templateRepo,omitempty" yaml:"templateRepo"`
	// Names of the file that will be processed by dinghy, by default is dinghyfile
	DinghyFilename                    string       `json:"dinghyFilename,omitempty" yaml:"dinghyFilename"`
	// Lock Dinghy pipelines
	AutoLockPipelines                 string       `json:"autoLockPipelines,omitempty" yaml:"autoLockPipelines"`
	// Overwrite deck baseUrl
	SpinnakerUIURL                    string       `json:"spinUIUrl,omitempty" yaml:"spinUIUrl"`
	// Github credentials path
	GitHubCredsPath                   string       `json:"githubCredsPath,omitempty" yaml:"githubCredsPath"`
	// Github token
	GitHubToken                       string       `json:"githubToken,omitempty" yaml:"githubToken"`
	// Github endpoint
	GithubEndpoint                    string       `json:"githubEndpoint,omitempty" yaml:"githubEndpoint"`
	// Gitlab Token
	GitLabToken                       string       `json:"gitlabToken,omitempty" yaml:"gitlabToken"`
	// Gitlanb api endpoint
	GitLabEndpoint                    string       `json:"gitlabEndpoint,omitempty" yaml:"gitlabEndpoint"`
	// Stash/Bitbucket credentials path
	StashCredsPath                    string       `json:"stashCredsPath,omitempty" yaml:"stashCredsPath"`
	// Stash/Bitbucket username
	StashUsername                     string       `json:"stashUsername,omitempty" yaml:"stashUsername"`
	// Stash/Bitbucket token
	StashToken                        string       `json:"stashToken,omitempty" yaml:"stashToken"`
	// Stash/Bitbucket api endpoint
	StashEndpoint                     string       `json:"stashEndpoint,omitempty" yaml:"stashEndpoint"`
	// Fiat service account
	FiatUser                          string       `json:"fiatUser,omitempty" yaml:"fiatUser"`
	Logging                           Logging      `json:"logging,omitempty" yaml:"logging"`
	// Secrets configuration
	Secrets                           Secrets      `json:"secrets,omitempty" yaml:"secrets"`
	// ParserFormat, supported formats are json, yaml and hcl
	ParserFormat                      string       `json:"parserFormat,omitempty" yaml:"parserFormat"`
	// Enable to process dinghyfiles from other branches
	RepoConfig                        []RepoConfig `json:"repoConfig,omitempty" yaml:"repoConfig"`
	// Spinnaker service endpoints
	spinnakerSupplied                 `mapstructure:",squash"`
	// Server configuration, default dinghy port is 8081
	Server                            server.ServerConfig `json:"server" yaml:"server"`
	// Configuration for certificates
	Http                              client.Config       `json:"http" yaml:"http"`
	// Webhook validations for repositories or orgs
	// More info here: https://docs.armory.io/docs/spinnaker-user-guides/using-dinghy/#webhook-secret-validation
	WebhookValidations                []WebhookValidation `json:"webhookValidations,omitempty" yaml:"webhookValidations"`
	// List of providers to check for webhook validation, currently only github supported
	// More info here: https://docs.armory.io/docs/spinnaker-user-guides/using-dinghy/#webhook-secret-validation
	WebhookValidationEnabledProviders []string            `json:"webhookValidationEnabledProviders,omitempty" yaml:"webhookValidationEnabledProviders"`
	// Repository template processing flag
	// More info here: https://docs.armory.io/docs/spinnaker-user-guides/using-dinghy/#repository-template-processing
	RepositoryRawdataProcessing       bool                `json:"repositoryRawdataProcessing,omitempty" yaml:"repositoryRawdataProcessing"`
	// This will be the TTL value to ger dinghyevents data
	LogEventTTLMinutes                time.Duration       `json:"LogEventTTLMinutes" yaml:"LogEventTTLMinutes"`
	// SQL configuration for dinghy
	SQL                               sqlconfig           `json:"sql,omitempty" yaml:"sql"`
}

type sqlconfig struct {
	// Enabled flag
	Enabled       bool   `json:"enabled,omitempty" yaml:"enabled"`
	// Database url
	BaseUrl       string `json:"baseUrl" yaml:"baseUrl"`
	// User
	User          string `json:"user" yaml:"user"`
	// Password
	Password      string `json:"password" yaml:"password"`
	// DB name
	DatabaseName  string `json:"databaseName" yaml:"databaseName"`
	// If this flag is enabled only events will be saved in database, redis will continue to be used for relationships
	EventLogsOnly bool   `json:"eventlogsOnly" yaml:"eventlogsOnly"`
}

type WebhookValidation struct {
	// Enabled flag
	Enabled                bool   `json:"enabled,omitempty" yaml:"enabled"`
	// Version control provider, only github supported at this moment
	VersionControlProvider string `json:"versionControlProvider,omitempty" yaml:"versionControlProvider"`
	// Organization
	Organization           string `json:"organization,omitempty" yaml:"organization"`
	// Repository
	Repo                   string `json:"repo,omitempty" yaml:"repo"`
	// Secret
	Secret                 string `json:"secret,omitempty" yaml:"secret"`
}

type spinnakerSupplied struct {
	// Orca service information
	Orca    spinnakerService `json:"orca,omitempty" yaml:"orca"`
	// Front50 service information
	Front50 spinnakerService `json:"front50,omitempty" yaml:"front50"`
	// Deck service information
	Deck    spinnakerService `json:"deck,omitempty" yaml:"deck"`
	// Echo service information
	Echo    spinnakerService `json:"echo,omitempty" yaml:"echo"`
	// Fiat service information
	Fiat    fiat             `json:"fiat,omitempty" yaml:"fiat"`
	// Redis service information, user will always be default
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
	// Vault configuration
	Vault secrets.VaultConfig `json:"vault" yaml:"vault"`
}

type RepoConfig struct {
	// Provider
	Provider string `json:"provider,omitempty" yaml:"provider"`
	// Repository
	Repo     string `json:"repo,omitempty" yaml:"repo"`
	// Branch
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
	if redacted.SQL.Password != "" {
		redacted.SQL.Password = "**REDACTED**"
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

// TraceExtract middleware extracts trace context from http headers following w3c trace context format
// and adds it to the request context
func (s *Settings) TraceExtract() func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqWithTraceContext := r
			next.ServeHTTP(w, reqWithTraceContext)
		})
	}
}
