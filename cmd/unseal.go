// Copyright © 2017 Lee Briggs <lee@leebriggs.co.uk>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/vault/api"
	v "github.com/jaxxstorm/hookpick/vault"

	"github.com/spf13/cobra"

	"sync"

	"github.com/jaxxstorm/hookpick/config"
	"github.com/jaxxstorm/hookpick/gpg"
)

// unsealCmd represents the unseal command
var unsealCmd = &cobra.Command{
	Use:   "unseal",
	Short: "Unseal the vaults using the key providers",
	Long: `Sends an unseal operationg to all vaults in the configuration file
using the key provided`,
	Run: func(cmd *cobra.Command, args []string) {

		allDCs := GetDatacenters()
		configHelper := NewConfigHelper(GetSpecificDatacenter, GetCaPath, GetProtocol, GetGpgKey)
		gpgHelper := gpg.NewGPGHelper(gpg.Decrypt)

		wg := sync.WaitGroup{}

		for _, dc := range allDCs {
			wg.Add(1)
			log.WithFields(log.Fields{
				"datacenter": dc.Name,
			}).Infoln("Start Vault unseal")
			go ProcessUnseal(&wg, &dc, configHelper, v.NewVaultHelper, gpgHelper, GetVaultKeys, UnsealHost)
		}
		wg.Wait()
	},
}

type VaultKeyGetter func(*config.Datacenter, ConfigKeyGetter, gpg.StringDecrypter) []string
type HostSubmitImpl func(*sync.WaitGroup, *v.VaultHelper, []string) bool

func ProcessUnseal(wg *sync.WaitGroup,
	dc *config.Datacenter,
	configHelper *ConfigHelper,
	vhGetter v.VaultHelperGetter,
	gpgHelper *gpg.GPGHelper,
	vaultKeysGetter VaultKeyGetter,
	unsealHost HostSubmitImpl) {

	defer wg.Done()

	specificDC := configHelper.GetDC()
	caPath := configHelper.GetCAPath()
	protocol := configHelper.GetURLScheme()

	if specificDC == dc.Name || specificDC == "" {

		vaultKeys := vaultKeysGetter(dc, configHelper.GetGPGKey, gpgHelper.Decrypt)

		hwg := sync.WaitGroup{}
		for _, host := range dc.Hosts {
			hwg.Add(1)
			vaultHelper := vhGetter(host.Name, caPath, protocol, host.Port, v.Status)
			go unsealHost(&hwg, vaultHelper, vaultKeys)
		}
		hwg.Wait()
	}
}

func GetVaultKeys(dc *config.Datacenter, gpgKeyGetter ConfigKeyGetter, keyDecrypter gpg.StringDecrypter) []string {
	var vaultKeys []string
	for _, key := range dc.Keys {
		gpg, gpgKey := gpgKeyGetter(key.Key, keyDecrypter)
		if gpg {
			vaultKeys = append(vaultKeys, gpgKey)
		} else {
			vaultKeys = append(vaultKeys, key.Key)
		}
	}

	return vaultKeys
}

func UnsealHost(wg *sync.WaitGroup, vaultHelper *v.VaultHelper, vaultKeys []string) bool {
	defer wg.Done()
	client, err := vaultHelper.GetVaultClient()
	if err != nil {
		log.WithFields(log.Fields{
			"host":  vaultHelper.HostName,
			"port":  vaultHelper.Port,
			"error": err,
		}).Error("Error creating Vault API Client")
	}

	// get the current status
	_, init := vaultHelper.GetStatus(client)
	if !init {
		// sad times, not ready to be unsealed
		log.WithFields(log.Fields{
			"host": vaultHelper.HostName,
		}).Error("Vault is not ready to be unsealed")
		return init
	}

	if len(vaultKeys) > 0 {
		var vaultStatus *api.SealStatusResponse
		for _, vaultKey := range vaultKeys {
			result, err := client.Sys().Unseal(vaultKey)
			// error while unsealing
			if err != nil {
				log.WithFields(log.Fields{
					"host": vaultHelper.HostName,
				}).Error("Error running unseal operation")
			}
			vaultStatus = result
		}

		// if it's still sealed, print the progress
		if vaultStatus.Sealed == true {
			log.WithFields(log.Fields{
				"host":      vaultHelper.HostName,
				"progress":  vaultStatus.Progress,
				"threshold": vaultStatus.T,
			}).Info("Unseal operation performed")
			// otherwise, tell us it's unsealed!
		} else {
			log.WithFields(log.Fields{
				"host":      vaultHelper.HostName,
				"progress":  vaultStatus.Progress,
				"threshold": vaultStatus.T,
			}).Info("Vault is unsealed!")
		}
	} else {
		log.WithFields(log.Fields{
			"host": vaultHelper.HostName,
		}).Error("No Key Provided")
	}

	return true
}

func init() {
	RootCmd.AddCommand(unsealCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// unsealCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// unsealCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
