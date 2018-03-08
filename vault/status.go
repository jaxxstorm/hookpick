package vault

import (
	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/vault/api"
)

// Status - Get vault status
func Status(client *api.Client) (bool, bool) {

	// statuses
	initStatus, err := client.Sys().InitStatus()

	if err != nil {
		log.WithFields(log.Fields{"host": client.Address()}).Error(err)
	}

	sealedStatus, err := client.Sys().SealStatus()

	if err != nil {
		log.WithFields(log.Fields{"host": client.Address()}).Error(err)
	}

	return sealedStatus.Sealed, initStatus
}
