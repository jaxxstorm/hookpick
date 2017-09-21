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
	v "github.com/jaxxstorm/unseal/vault"
	"github.com/spf13/cobra"

	log "github.com/Sirupsen/logrus"

	//"github.com/acidlemon/go-dumper"

	"sync"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get the status of all vaults",
	Long: `Will call the status API endpoint for all vaults
specified in the configuration file`,
	Run: func(cmd *cobra.Command, args []string) {

		datacenters := getDatacenters()

		caPath := getCaPath()

		var wg sync.WaitGroup

		for _, d := range datacenters {
			for _, h := range d.Hosts {

				// set hostnames for waitgroup
				hostName := h.Name
				hostPort := h.Port

				wg.Add(1)

				go func(hostName string, hostPort int) {
					defer wg.Done()

					client, err := v.VaultClient(hostName, hostPort, caPath)

					if err != nil {
						log.WithFields(log.Fields{"host": hostName}).Error("Error creating vault client: ", err)
					}

					// get the seal status
					result, err := client.Sys().SealStatus()

					if err != nil {
						log.WithFields(log.Fields{"host": hostName}).Error("Error getting seal status: ", err)
					} else {
						// only check the seal status if we have a client
						if result.Sealed == true {
							log.WithFields(log.Fields{"host": hostName, "progress": result.Progress, "threshold": result.T}).Error("Vault is sealed!")
						} else {
							log.WithFields(log.Fields{"host": hostName, "progress": result.Progress, "threshold": result.T}).Info("Vault is unsealed!")
						}
					}
				}(hostName, hostPort)
			}
		}
		wg.Wait()

	},
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
