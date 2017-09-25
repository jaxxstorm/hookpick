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
	"github.com/hashicorp/vault/api"
	"sync"
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

		datacenters := getDatacenters()
		caPath := getCaPath()

		var wg sync.WaitGroup

		// loop through datacenters
		for _, d := range datacenters {
			datacenter := getSpecificDatacenter()
			if datacenter == d.Name || datacenter == "" {

				// loop through hosts
				for _, h := range d.Hosts {
					hostName := h.Name
					hostPort := h.Port

					wg.Add(1)

					// set up vault client
					go func(hostName string, hostPort int) {
						client, err := v.VaultClient(hostName, hostPort, caPath)

						if err != nil {
							log.WithFields(log.Fields{"host": hostName, "port": hostPort}).Error(err)
						}

						// check init status
						sealed, init := v.Status(client)

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
									log.WithFields(log.Fields{"host": hostName, "shares": rekeyResult.N, "threshold": rekeyResult.T, "nonce": rekeyResult.Nonce}).Info("Rekey Started. Please supply your keys.")
								}
							}
						}
					}(hostName, hostPort)
				}

			}
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

	},
}

var rekeyStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Retrieves the current status of a rekey",
	Long: `Retrieves the current status of a rekey process
from all the specified Vault servers`,
	Run: func(cmd *cobra.Command, args []string) {

		datacenters := getDatacenters()
		caPath := getCaPath()

		var wg sync.WaitGroup

		for _, d := range datacenters {
			datacenter := getSpecificDatacenter()
			if datacenter == d.Name || datacenter == "" {
				for _, h := range d.Hosts {
					hostName := h.Name
					hostPort := h.Port

					wg.Add(1)
					go func(hostName string, hostPort int) {
						defer wg.Done()
						client, err := v.VaultClient(hostName, hostPort, caPath)

						if err != nil {
							log.WithFields(log.Fields{"host": hostName, "port": hostPort}).Error(err)
						}

						// check init status
						sealed, init := v.Status(client)

						if init == true && sealed == false {
							result, _ := client.Sys().Leader()
							// if we are the leader start the rekey
							if result.IsSelf == true {
								rekeyStatus, err := client.Sys().RekeyStatus()

								if err != nil {
									log.WithFields(log.Fields{"host": hostName, "port": hostPort}).Error(err)
								}
								if rekeyStatus.Started {
									log.WithFields(log.Fields{"host": hostName, "shares": rekeyStatus.N, "threshold": rekeyStatus.T, "nonce": rekeyStatus.Nonce}).Info("Rekey has been started")
								} else {
									log.WithFields(log.Fields{"host": hostName}).Info("Rekey not started")
								}
							}
						}
					}(hostName, hostPort)
				}
			}
		}
		wg.Wait()
	},
}

func init() {
	RootCmd.AddCommand(rekeyCmd)
	rekeyCmd.AddCommand(initCmd)
	rekeyCmd.AddCommand(submitCmd)
	rekeyCmd.AddCommand(rekeyStatusCmd)

	initCmd.Flags().IntVarP(&shares, "shares", "s", 0, "The number of secret shares to init the rekey with")
	initCmd.Flags().IntVarP(&threshold, "threshold", "t", 0, "The secret threshold to init the rekey with")

}
