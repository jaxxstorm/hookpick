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
	v "github.com/jaxxstorm/hookpick/vault"
	"github.com/spf13/cobra"

	log "github.com/Sirupsen/logrus"

	"github.com/jaxxstorm/hookpick/config"
	"sync"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get the status of all vaults",
	Long: `Queries the status of all vaults
specified in the configuration file`,
	Run: func(cmd *cobra.Command, args []string) {

		datacenters := GetDatacenters()
		configHelper := NewConfigHelper(GetSpecificDatacenter, GetCaPath, GetProtocol, GetGpgKey)
		wg := sync.WaitGroup{}

		for _, dc := range datacenters {
			wg.Add(1)
			go ProcessStatus(&wg, &dc, configHelper, v.NewVaultHelper, GetHostStatus)
		}
		wg.Wait()
	},
}

type HostImpl func(*sync.WaitGroup, *v.VaultHelper)

func ProcessStatus(wg *sync.WaitGroup,
	dc *config.Datacenter,
	configHelper *ConfigHelper,
	vhGetter v.VaultHelperGetter,
	hostStatusGetter HostImpl) {

	defer wg.Done()

	specificDC := configHelper.GetDC()
	caPath := configHelper.GetCAPath()
	protocol := configHelper.GetURLScheme()

	if specificDC == dc.Name || specificDC == "" {

		hwg := sync.WaitGroup{}
		for _, host := range dc.Hosts {
			hwg.Add(1)
			vaultHelper := vhGetter(host.Name, caPath, protocol, host.Port, v.Status)
			go hostStatusGetter(&hwg, vaultHelper)
		}
		hwg.Wait()
	}
}

func GetHostStatus(wg *sync.WaitGroup, vaultHelper *v.VaultHelper) {
	// set hostnames for waitgroup

	defer wg.Done()

	client, err := vaultHelper.GetVaultClient()

	if err != nil {
		log.WithFields(log.Fields{"host": vaultHelper.HostName}).Error("Error creating vault client: ", err)
	}

	// get the seal status
	result, err := client.Sys().SealStatus()

	if err != nil {
		log.WithFields(log.Fields{"host": vaultHelper.HostName}).Error("Error getting seal status: ", err)
	} else {
		// only check the seal status if we have a client
		if result.Sealed == true {
			log.WithFields(log.Fields{"host": vaultHelper.HostName, "progress": result.Progress, "threshold": result.T}).Error("Vault is sealed!")
		} else {
			log.WithFields(log.Fields{"host": vaultHelper.HostName, "progress": result.Progress, "threshold": result.T}).Info("Vault is unsealed!")
		}
	}
}


func init() {
	RootCmd.AddCommand(statusCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
