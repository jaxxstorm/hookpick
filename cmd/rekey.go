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
	"github.com/spf13/cobra"

	v "github.com/jaxxstorm/hookpick/vault"

	//"github.com/acidlemon/go-dumper"
	"sync"

	"github.com/hashicorp/vault/api"
	"github.com/jaxxstorm/hookpick/config"
	"github.com/jaxxstorm/hookpick/gpg"
)

var shares int
var threshold int

// rekeyCmd represents the rekey command
var rekeyCmd = &cobra.Command{
	Use:   "rekey",
	Short: "Runs rekey operations against Vault servers",
	Long: `Runs rekey operations against all Vault servers
or specified Vault servers in the configuration file`,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Starts the rekey operation on specified Vault server",
	Long: `Initialises a rekey against specified Vault servers
and returns the client nonce needed for other rekey operators`,
	Run: func(cmd *cobra.Command, args []string) {

		if threshold == 0 {
			log.Fatal("Please specify the secret threshold: See --help")
		}

		if shares == 0 {
			log.Fatal("Please specify the secret shares: See --help")
		}

		allDCs := GetDatacenters()
		configHelper := NewConfigHelper(GetSpecificDatacenter, GetCaPath, GetProtocol, GetGpgKey)

		wg := sync.WaitGroup{}
		// loop through datacenters
		for _, dc := range allDCs {
			wg.Add(1)
			go ProcessRekey(&wg, dc, configHelper, v.NewVaultHelper, HostRekeyInit)
		}
		wg.Wait()
	},
}

var submitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submits your key to the rekey command",
	Long: `Submits your unseal key to the rekey process
and progresses the rekey`,
	Run: func(cmd *cobra.Command, args []string) {
		allDCs := GetDatacenters()
		configHelper := NewConfigHelper(GetSpecificDatacenter, GetCaPath, GetProtocol, GetGpgKey)
		gpgHelper := gpg.NewGPGHelper(gpg.Decrypt)

		wg := sync.WaitGroup{}

		for _, dc := range allDCs {
			wg.Add(1)
			go ProcessRekeySubmit(&wg, dc, configHelper, v.NewVaultHelper, gpgHelper, GetVaultKeys, HostRekeySubmit)
		}
		wg.Wait()
	},
}

var rekeyStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Retrieves the current status of a rekey",
	Long: `Retrieves the current status of a rekey process
from all the specified Vault servers`,
	Run: func(cmd *cobra.Command, args []string) {

		allDCs := GetDatacenters()
		configHelper := NewConfigHelper(GetSpecificDatacenter, GetCaPath, GetProtocol, GetGpgKey)

		wg := sync.WaitGroup{}

		for _, dc := range allDCs {
			wg.Add(1)
			go ProcessRekey(&wg, dc, configHelper, v.NewVaultHelper, HostRekeyStatus)
		}
		wg.Wait()
	},
}

func ProcessRekey(wg *sync.WaitGroup,
	dc config.Datacenter,
	configHelper *ConfigHelper,
	vhGetter v.VaultHelperGetter,
	hostRekeyInit HostImpl) {
	defer wg.Done()

	specificDC := configHelper.GetDC()
	caPath := configHelper.GetCAPath()
	protocol := configHelper.GetURLScheme()

	if specificDC == dc.Name || specificDC == "" {

		hwg := sync.WaitGroup{}
		for _, host := range dc.Hosts {
			hwg.Add(1)
			vaultHelper := vhGetter(host.Name, caPath, protocol, host.Port, v.Status)
			go hostRekeyInit(&hwg, vaultHelper)
		}
		hwg.Wait()
	}
}

func ProcessRekeySubmit(wg *sync.WaitGroup,
	dc config.Datacenter,
	configHelper *ConfigHelper,
	vhGetter v.VaultHelperGetter,
	gpgHelper *gpg.GPGHelper,
	vaultKeysGetter VaultKeyGetter,
	submitHostRekey HostSubmitImpl) {
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
			go submitHostRekey(&hwg, vaultHelper, vaultKeys)
		}
		hwg.Wait()
	}
}

func HostRekeyInit(wg *sync.WaitGroup, vaultHelper *v.VaultHelper) {
	defer wg.Done()
	client, err := vaultHelper.GetVaultClient()

	if err != nil {
		log.WithFields(log.Fields{"host": vaultHelper.HostName, "port": vaultHelper.Port}).Error(err)
	}

	// check init status
	sealed, init := vaultHelper.GetStatus(client)

	if init == true && sealed == false {
		// get the current leader to operate on
		result, _ := client.Sys().Leader()
		// if we are the leader start the rekey
		if result.IsSelf == true {
			rekeyResult, err := client.Sys().RekeyInit(&api.RekeyInitRequest{SecretShares: shares, SecretThreshold: threshold})
			if err != nil {
				log.Error("Rekey init error ", err)
			}
			if rekeyResult.Started {
				log.WithFields(log.Fields{
					"host":      vaultHelper.HostName,
					"shares":    rekeyResult.N,
					"threshold": rekeyResult.T,
					"nonce":     rekeyResult.Nonce,
				}).Info("Rekey Started. Please supply your keys.")
			}
		}
	}
}

func HostRekeyStatus(wg *sync.WaitGroup, vaultHelper *v.VaultHelper) {
	defer wg.Done()
	client, err := vaultHelper.GetVaultClient()

	if err != nil {
		log.WithFields(log.Fields{"host": vaultHelper.HostName, "port": vaultHelper.Port}).Error(err)
	}

	// check init status
	sealed, init := vaultHelper.GetStatus(client)

	if init == true && sealed == false {
		result, _ := client.Sys().Leader()
		// if we are the leader start the rekey
		if result.IsSelf == true {
			rekeyStatus, err := client.Sys().RekeyStatus()

			if err != nil {
				log.WithFields(log.Fields{"host": vaultHelper.HostName, "port": vaultHelper.Port}).Error(err)
			}
			if rekeyStatus.Started {
				log.WithFields(log.Fields{"host": vaultHelper.HostName, "shares": rekeyStatus.N, "threshold": rekeyStatus.T, "nonce": rekeyStatus.Nonce, "progress": rekeyStatus.Progress, "required": rekeyStatus.Required}).Info("Rekey has been started")
			} else {
				log.WithFields(log.Fields{"host": vaultHelper.HostName}).Info("Rekey not started")
			}
		}
	}

}

func HostRekeySubmit(wg *sync.WaitGroup, vaultHelper *v.VaultHelper, vaultKeys []string) bool {
	defer wg.Done()
	client, err := vaultHelper.GetVaultClient()
	if err != nil {
		log.WithFields(log.Fields{
			"host": vaultHelper.HostName,
			"port": vaultHelper.Port,
		}).Error(err)
	}

	// check init status
	sealed, init := vaultHelper.GetStatus(client)

	if init == true && sealed == false {
		result, _ := client.Sys().Leader()
		// if we are the leader start the rekey
		if result.IsSelf == true {
			rekeyStatus, err := client.Sys().RekeyStatus()
			if err != nil {
				log.WithFields(log.Fields{
					"host": vaultHelper.HostName,
					"port": vaultHelper.Port,
				}).Error(err)
				return false
			}

			if rekeyStatus.Started {
				for _, vaultKey := range vaultKeys {
					rekeyUpdate, err := client.Sys().RekeyUpdate(vaultKey, rekeyStatus.Nonce)
					if err != nil {
						log.WithFields(log.Fields{
							"host": vaultHelper.HostName,
							"port": vaultHelper.Port,
						}).Error(err)
						continue
					}

					if rekeyUpdate.Complete {

						var outputKeys []string
						var outputPgps []string

						log.WithFields(log.Fields{
							"host": vaultHelper.HostName,
						}).Info("Rekey Complete")

						for _, key := range rekeyUpdate.KeysB64 {
							outputKeys = append(outputKeys, key)
						}

						for _, pgp := range rekeyUpdate.PGPFingerprints {
							outputPgps = append(outputPgps, pgp)
						}

						for _, outputPgp := range outputPgps {
							log.WithFields(log.Fields{
								"PGP Fingerprint": outputPgp,
							}).Info("New Key Generated")
						}

						for _, outputKey := range outputKeys {
							log.WithFields(log.Fields{
								"Key": outputKey,
							}).Info("New Key Generated")
						}

						break
					} else {
						newRekeyStatus, err := client.Sys().RekeyStatus()
						if err != nil {
							log.WithFields(log.Fields{
								"host": vaultHelper.HostName,
								"port": vaultHelper.Port,
							}).Error(err)
						}
						log.WithFields(log.Fields{
							"host":      vaultHelper.HostName,
							"shares":    newRekeyStatus.N,
							"threshold": newRekeyStatus.T,
							"nonce":     newRekeyStatus.Nonce,
							"progress":  newRekeyStatus.Progress,
							"required":  newRekeyStatus.Required,
						}).Info("Key submitted")
					}
				}
			} else {
				log.WithFields(log.Fields{"host": vaultHelper.HostName}).Info("Rekey not started")
			}
		}
	}
	return true
}

func init() {
	RootCmd.AddCommand(rekeyCmd)
	rekeyCmd.AddCommand(initCmd)
	rekeyCmd.AddCommand(submitCmd)
	rekeyCmd.AddCommand(rekeyStatusCmd)

	initCmd.Flags().IntVarP(&shares, "shares", "s", 0, "The number of secret shares to init the rekey with")
	initCmd.Flags().IntVarP(&threshold, "threshold", "t", 0, "The secret threshold to init the rekey with")

}
