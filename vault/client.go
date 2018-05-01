package vault

import (
	vaultapi "github.com/hashicorp/vault/api"
	"net/url"
	"strings"
)

type VaultHelper struct {
	HostName  string
	Port      string
	CAPath    string
	Protocol  string
	GetStatus VaultStatusGetter
}

type VaultHelperGetter func(host, certpath, protocol, port string, sg VaultStatusGetter) *VaultHelper

func NewVaultHelper(host, certpath, protocol, port string, sg VaultStatusGetter) *VaultHelper {
	return &VaultHelper{
		HostName:  host,
		Port:      port,
		CAPath:    certpath,
		Protocol:  protocol,
		GetStatus: sg,
	}
}

// VaultClient Create a Vault Client
func (helper *VaultHelper) GetVaultClient() (*vaultapi.Client, error) {
	urlhoststrings := []string{helper.HostName, helper.Port}
	hostUrl := url.URL{
		Host:   strings.Join(urlhoststrings, ":"),
		Scheme: helper.Protocol,
	}

	// create a vault config
	config := &vaultapi.Config{Address: hostUrl.String()}

	// read in any environment variables that might be set
	if err := config.ReadEnvironment(); err != nil {
		return nil, err
	}

	// Set the CA path, if it's present
	if err := config.ConfigureTLS(&vaultapi.TLSConfig{CAPath: helper.CAPath}); err != nil {
		return nil, err
	}

	// create the client
	client, err := vaultapi.NewClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
