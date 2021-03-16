package secret

import (
	"context"
	"github.com/armory/dinghy/pkg/settings/lighthouse"
	"github.com/armory/go-yaml-tools/pkg/secrets"
)

type SecretHandler struct {
}

func NewSecretHandler(s lighthouse.Secrets) (*SecretHandler, error) {

	if (lighthouse.Secrets{}) != s {
		if (secrets.VaultConfig{}) != s.Vault {
			if err := secrets.RegisterVaultConfig(s.Vault); err != nil {
				return nil, err
			}
		}
	}

	return &SecretHandler{}, nil
}

func (sh *SecretHandler) Decrypt(ctx context.Context, config *lighthouse.Settings) error {
	decrypter, err := secrets.NewDecrypter(ctx, config.GitHubToken)
	if err != nil {
		return err
	}
	secret, err := decrypter.Decrypt()
	if err != nil {
		return err
	}
	config.GitHubToken = secret

	decrypter, err = secrets.NewDecrypter(ctx, config.StashToken)
	if err != nil {
		return err
	}
	secret, err = decrypter.Decrypt()
	if err != nil {
		return err
	}
	config.StashToken = secret

	decrypter, err = secrets.NewDecrypter(ctx, config.SQL.Password)
	if err != nil {
		return err
	}
	secret, err = decrypter.Decrypt()
	if err != nil {
		return err
	}
	config.SQL.Password = secret
	return nil
}
