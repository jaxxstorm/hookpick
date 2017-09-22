package vault

import (
	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/vault/api"
)

func InitStatus(client *api.Client) Status {

	// statuses
	init, err := client.Sys().InitStatus()

	if err != nil {
		log.WithFields(log.Fields{"host": client.Address()}).Error(err)
		return Status{
			Ready:  false,
			Reason: "Error while retrieving initstatus",
		}
	}
	if init == false {
		return Status{
			Ready:  false,
			Reason: "Vault is not initialized",
		}
	}

	seal, err := client.Sys().SealStatus()

	if err != nil {
		log.WithFields(log.Fields{"host": client.Address()}).Error(err)
	}

	if seal.Sealed != true {
		return Status{
			Ready:  true,
			Reason: "Vault is already unsealed",
		}
	}

	return Status{
		Ready: true,
	}

}

func SealedStatus(client *api.Client) bool {
	status, err := client.Sys().SealStatus()

	if err != nil {
		log.WithFields(log.Fields{"host": client.Address()}).Error(err)
	}

	return status.Sealed

}
