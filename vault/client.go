package vault

import (
	"fmt"
	"github.com/hashicorp/vault/api"
)

func VaultClient(hostName string, hostPort int, caPath string) (*api.Client, error) {

	// format the URL with the passed host and port
	protocol := "https"
	if caPath == "" {
		protocol = "http"
	}
	url := fmt.Sprintf("%s://%s:%v", protocol, hostName, hostPort)

	// create a vault config
	config := &api.Config{Address: url}

	// read in any environment variables that might be set
	if err := config.ReadEnvironment(); err != nil {
		return nil, err
	}

	// Set the CA path, if it's present
	if err := config.ConfigureTLS(&api.TLSConfig{CAPath: caPath}); err != nil {
		return nil, err
	}

	// create the client
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil

}
