package secrets

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
)

type VaultConfig struct {
	Enabled      bool   `json:"enabled" yaml:"enabled"`
	Url          string `json:"url" yaml:"url"`
	AuthMethod   string `json:"authMethod" yaml:"authMethod"`
	Role         string `json:"role" yaml:"role"`
	Path         string `json:"path" yaml:"path"`
	Username     string `json:"username" yaml:"username"`
	Password     string `json:"password" yaml:"password"`
	UserAuthPath string `json:"userAuthPath" yaml:"userAuthPath"`
	Namespace    string `json:"namespace" yaml:"namespace"`
	Token        string // no struct tags for token
}

type VaultSecret struct {
}

type VaultDecrypter struct {
	engine        string
	path          string
	key           string
	base64Encoded string
	isFile        bool
	vaultConfig   VaultConfig
	tokenFetcher  TokenFetcher
}

type VaultClient interface {
	Write(path string, data map[string]interface{}) (*api.Secret, error)
	Read(path string) (*api.Secret, error)
}

func RegisterVaultConfig(vaultConfig VaultConfig) error {
	if err := validateVaultConfig(vaultConfig); err != nil {
		return fmt.Errorf("vault configuration error - %s", err)
	}

	Engines["vault"] = func(ctx context.Context, isFile bool, params string) (Decrypter, error) {
		vd := &VaultDecrypter{isFile: isFile, vaultConfig: vaultConfig}
		if err := vd.parseSyntax(params); err != nil {
			return nil, err
		}
		if err := vd.setTokenFetcher(); err != nil {
			return nil, err
		}
		return vd, nil
	}
	return nil
}

type TokenFetcher interface {
	fetchToken(client VaultClient) (string, error)
}

type EnvironmentVariableTokenFetcher struct{}

func (e EnvironmentVariableTokenFetcher) fetchToken(client VaultClient) (string, error) {
	token := os.Getenv("VAULT_TOKEN")
	if token == "" {
		return "", fmt.Errorf("VAULT_TOKEN environment variable not set")
	}
	return token, nil
}

type UserPassTokenFetcher struct {
	username     string
	password     string
	userAuthPath string
}

func (u UserPassTokenFetcher) fetchToken(client VaultClient) (string, error) {
	data := map[string]interface{}{
		"password": u.password,
	}
	loginPath := "auth/" + u.userAuthPath + "/login/" + u.username

	log.Infof("logging into vault with USERPASS auth at: %s", loginPath)
	secret, err := client.Write(loginPath, data)
	if err != nil {
		return handleLoginErrors(err)
	}

	return secret.Auth.ClientToken, nil
}

type KubernetesServiceAccountTokenFetcher struct {
	role string
	path string
	fileReader fileReader
}

// define a file reader function so we can test kubernetes auth
type fileReader func(string) ([]byte, error)

func (k KubernetesServiceAccountTokenFetcher) fetchToken(client VaultClient) (string, error) {
	tokenBytes, err := k.fileReader("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		return "", fmt.Errorf("error reading service account token: %s", err)
	}
	data := map[string]interface{}{
		"role": k.role,
		"jwt":  string(tokenBytes),
	}
	loginPath := "auth/" + k.path + "/login"

	log.Infof("logging into vault with KUBERNETES auth at: %s", loginPath)
	secret, err := client.Write(loginPath, data)
	if err != nil {
		return handleLoginErrors(err)
	}

	return secret.Auth.ClientToken, nil
}

func handleLoginErrors(err error) (string, error) {
	if _, ok := err.(*json.SyntaxError); ok {
		// some connection errors aren't properly caught, and the vault client tries to parse <nil>
		return "", fmt.Errorf("error fetching secret from vault - check connection to the server")
	}
	return "", fmt.Errorf("error logging into vault: %s", err)
}

func (decrypter *VaultDecrypter) setTokenFetcher() error {
	var tokenFetcher TokenFetcher

	switch decrypter.vaultConfig.AuthMethod {
	case "TOKEN":
		tokenFetcher = EnvironmentVariableTokenFetcher{}
	case "KUBERNETES":
		tokenFetcher = KubernetesServiceAccountTokenFetcher{
			role: decrypter.vaultConfig.Role,
			path: decrypter.vaultConfig.Path,
			fileReader: ioutil.ReadFile,
		}
	case "USERPASS":
		tokenFetcher = UserPassTokenFetcher{
			username:     decrypter.vaultConfig.Username,
			password:     decrypter.vaultConfig.Password,
			userAuthPath: decrypter.vaultConfig.UserAuthPath,
		}
	default:
		return fmt.Errorf("unknown Vault secrets auth method: %q", decrypter.vaultConfig.AuthMethod)
	}

	decrypter.tokenFetcher = tokenFetcher
	return nil
}

func (decrypter *VaultDecrypter) Decrypt() (string, error) {
	if decrypter.vaultConfig.Token == "" {
		err := decrypter.setToken()
		if err != nil {
			return "", err
		}
	}
	client, err := decrypter.getVaultClient()
	if err != nil {
		return "", err
	}
	secret, err := decrypter.fetchSecret(client)
	if err != nil && strings.Contains(err.Error(), "403") {
		// get new token and retry in case our saved token is no longer valid
		err := decrypter.setToken()
		if err != nil {
			return "", err
		}
		secret, err = decrypter.fetchSecret(client)
	}
	if err != nil {
		return "", err
	}
	if decrypter.IsFile() {
		return ToTempFile([]byte(secret))
	}
	return secret, nil
}

func (v *VaultDecrypter) IsFile() bool {
	return v.isFile
}

func (v *VaultDecrypter) parseSyntax(params string) error {
	tokens := strings.Split(params, "!")
	for _, element := range tokens {
		kv := strings.Split(element, ":")
		if len(kv) == 2 {
			switch kv[0] {
			case "e":
				v.engine = kv[1]
			case "p", "n":
				v.path = kv[1]
			case "k":
				v.key = kv[1]
			case "b":
				v.base64Encoded = kv[1]
			}
		}
	}

	if v.engine == "" {
		return fmt.Errorf("secret format error - 'e' for engine is required")
	}
	if v.path == "" {
		return fmt.Errorf("secret format error - 'p' for path is required (replaces deprecated 'n' param)")
	}
	if v.key == "" {
		return fmt.Errorf("secret format error - 'k' for key is required")
	}
	return nil
}

func validateVaultConfig(vaultConfig VaultConfig) error {
	if (VaultConfig{}) == vaultConfig {
		return fmt.Errorf("vault secrets not configured in service profile yaml")
	}
	if vaultConfig.Enabled == false {
		return fmt.Errorf("vault secrets disabled")
	}
	if vaultConfig.Url == "" {
		return fmt.Errorf("vault url required")
	}
	if vaultConfig.AuthMethod == "" {
		return fmt.Errorf("auth method required")
	}

	switch vaultConfig.AuthMethod {
	case "TOKEN":
		if vaultConfig.Token == "" {
			envToken := os.Getenv("VAULT_TOKEN")
			if envToken == "" {
				return fmt.Errorf("VAULT_TOKEN environment variable not set")
			}
		}
	case "KUBERNETES":
		if vaultConfig.Path == "" || vaultConfig.Role == "" {
			return fmt.Errorf("path and role both required for KUBERNETES auth method")
		}
	case "USERPASS":
		if vaultConfig.Username == "" || vaultConfig.Password == "" || vaultConfig.UserAuthPath == "" {
			return fmt.Errorf("username, password and userAuthPath are required for USERPASS auth method")
		}
	default:
		return fmt.Errorf("unknown Vault secrets auth method: %q", vaultConfig.AuthMethod)
	}

	return nil
}

func (decrypter *VaultDecrypter) setToken() error {
	client, err := decrypter.getVaultClient()
	if err != nil {
		return err
	}
	token, err := decrypter.tokenFetcher.fetchToken(client)
	if err != nil {
		return fmt.Errorf("error fetching vault token - %s", err)
	}
	decrypter.vaultConfig.Token = token
	return nil
}

func (decrypter *VaultDecrypter) getVaultClient() (*api.Logical, error) {
	client, err := decrypter.newAPIClient()
	if err != nil {
		return nil, err
	}
	return client.Logical(), nil
}

func (decrypter *VaultDecrypter) newAPIClient() (*api.Client, error) {
	client, err := api.NewClient(&api.Config{
		Address: decrypter.vaultConfig.Url,
	})
	if err != nil {
		return nil, fmt.Errorf("error fetching vault client: %s", err)
	}
	if decrypter.vaultConfig.Namespace != "" {
		client.SetNamespace(decrypter.vaultConfig.Namespace)
	}
	if decrypter.vaultConfig.Token != "" {
		client.SetToken(decrypter.vaultConfig.Token)
	}
	return client, nil
}


func (decrypter *VaultDecrypter) fetchSecret(client VaultClient) (string, error) {
	path := decrypter.engine + "/" + decrypter.path
	log.Infof("attempting to read secret at KV v1 path: %s", path)
	secretMapping, v1err := client.Read(path)
	if v1err != nil {
		if _, ok := v1err.(*json.SyntaxError); ok {
			// some connection errors aren't properly caught, and the vault client tries to parse <nil>
			return "", fmt.Errorf("error fetching secret from vault - check connection to the server: %s",
				decrypter.vaultConfig.Url)
		}
	}

	var v2err error
	if containsRetryableError(v1err, secretMapping) {
		// try again using K/V v2 path
		path = decrypter.engine + "/data/" + decrypter.path
		log.Infof("attempting to read secret at KV v2 path: %s", path)
		secretMapping, v2err = client.Read(path)
	}

	if v2err != nil {
		log.Errorf("error reading secret at KV v1 path and KV v2 path")
		log.Errorf("KV v1 error: %s", v1err)
		log.Errorf("KV v2 error: %s", v2err)
		return "", fmt.Errorf("error fetching secret from vault")
	}

	return decrypter.parseResults(secretMapping)
}

func containsRetryableError(err error, secret *api.Secret) bool {
	if err != nil || secret == nil {
		return true
	}
	warnings := secret.Warnings
	for _, w := range warnings {
		switch {
		case strings.Contains(w, "Invalid path for a versioned K/V secrets engine"):
			return true
		}
	}
	return false
}

func (decrypter *VaultDecrypter) parseResults(secretMapping *api.Secret) (string, error) {
	if secretMapping == nil {
		return "", fmt.Errorf("couldn't find vault path %s under engine %s", decrypter.path, decrypter.engine)
	}

	mapping := secretMapping.Data
	if data, ok := mapping["data"]; ok { // one more nesting of "data" if using K/V v2
		if submap, ok := data.(map[string]interface{}); ok {
			mapping = submap
		}
	}

	decrypted, ok := mapping[decrypter.key].(string)
	if !ok {
		return "", fmt.Errorf("key %q not found at engine: %s, path: %s", decrypter.key, decrypter.engine, decrypter.path)
	}
	log.Debugf("successfully fetched secret")
	return decrypted, nil
}

func DecodeVaultConfig(vaultYaml map[interface{}]interface{}) (*VaultConfig, error) {
	var cfg VaultConfig
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           &cfg,
		WeaklyTypedInput: true,
	})
	if err != nil {
		return nil, err
	}

	if err := decoder.Decode(vaultYaml); err != nil {
		return nil, err
	}

	return &cfg, nil
}
