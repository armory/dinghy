package secret

import (
	"context"
	"github.com/armory/dinghy/pkg/settings/global"
	"github.com/armory/go-yaml-tools/pkg/secrets"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewSecretHandler(t *testing.T) {
	i, err := NewSecretHandler(global.Secrets{})

	assert.Nil(t, err)
	assert.NotNil(t, i)
}

func TestNewSecretHandlerWithVault(t *testing.T) {
	i, err := NewSecretHandler(global.Secrets{Vault: secrets.VaultConfig{
		Enabled:    true,
		Url:        "http://locahost:8200/",
		AuthMethod: "TOKEN",
		Token:      "xxxxxxxxx",
	}})

	_, ok := secrets.Engines["vault"]

	assert.Nil(t, err)
	assert.NotNil(t, i)
	assert.Equal(t, true, ok)
}

func TestSecretHandler_Decrypt(t *testing.T) {
	i, err := NewSecretHandler(global.Secrets{})
	assert.Nil(t, err)
	assert.NotNil(t, i)

	RegisterTestConfig()

	config := &global.Settings{
		GitHubToken: "encrypted:test!GitHubToken",
		StashToken:  "encrypted:test!StashToken",
		SQL: global.Sqlconfig{
			Password: "encrypted:test!Password",
		},
	}
	err = i.Decrypt(context.TODO(), config)
	assert.Nil(t, err)

	assert.Equal(t, "supersecretValue", config.GitHubToken)
	assert.Equal(t, "supersecretValue", config.StashToken)
	assert.Equal(t, "supersecretValue", config.SQL.Password)
}

func RegisterTestConfig() {
	secrets.Engines["test"] = func(ctx context.Context, isFile bool, params string) (secrets.Decrypter, error) {
		return &TestDecrypter{}, nil
	}
}

type TestDecrypter struct{}

func (*TestDecrypter) Decrypt() (string, error) {
	return "supersecretValue", nil
}

func (v *TestDecrypter) IsFile() bool {
	return false
}
