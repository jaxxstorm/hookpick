// Copyright © 2017 Lee Briggs
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
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/spf13/viper"

	log "github.com/Sirupsen/logrus"
	"github.com/jaxxstorm/hookpick/config"
	g "github.com/jaxxstorm/hookpick/gpg"
)

var cfgFile string

var datacenter string

var datacenters []config.Datacenter

// Version : This is for the Version command
var Version string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "hookpick",
	Short: "A tool to manage Vault clusters",
	Long: `Easily unseal, rekey and init multiple Vault servers in a large,
distributed environment`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {
	Version = version
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.hookpick.yaml)")
	RootCmd.PersistentFlags().StringVarP(&datacenter, "datacenter", "d", "", "datacenter to operate on")
	viper.BindPFlag("datacenter", RootCmd.PersistentFlags().Lookup("datacenter"))

	if os.Getenv("VAULT_ADDR") != "" {
		log.Warning("Warning VAULT_ADDR environment variable is set. This will override the hostname in your config file, it's probably not what you want")
	}
}

type ConfigStringGetter func() string
type ConfigKeyGetter func(string, g.StringDecrypter) (bool, string)

type ConfigHelper struct {
	GetDC        ConfigStringGetter
	GetCAPath    ConfigStringGetter
	GetURLScheme ConfigStringGetter
	GetGPGKey    ConfigKeyGetter
}

func NewConfigHelper(dcgetter, cagetter, schemegetter ConfigStringGetter, gpgkeygetter ConfigKeyGetter) *ConfigHelper {
	return &ConfigHelper{
		GetDC:        dcgetter,
		GetCAPath:    cagetter,
		GetURLScheme: schemegetter,
		GetGPGKey:    gpgkeygetter,
	}
}

func GetDatacenters() []config.Datacenter {

	err := viper.UnmarshalKey("datacenters", &datacenters)

	if err != nil {
		log.Error("Unable to read hosts key in config file: %s", err)
	}

	return datacenters

}

func GetCaPath() string {

	return viper.GetString("capath")

}

func GetGpgKey(key string, keyDecrypt g.StringDecrypter) (bool, string) {

	gpg := viper.GetBool("gpg")
	var vaultKey string
	var err error

	if gpg == true {
		vaultKey, err = keyDecrypt(key)
		if err != nil {
			log.Fatal("GPG Decryption Error: ", err)
		}
	} else {

		vaultKey = ""
	}

	return gpg, vaultKey

}

func GetSpecificDatacenter() string {

	return viper.GetString("datacenter")

}

func GetProtocol() string {

	viper.SetDefault("protocol", "https")

	return viper.GetString("protocol")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".hookpick") // name of config file (without extension)
		viper.AddConfigPath("$HOME")     // adding home directory as first search path
		viper.AddConfigPath(".")
		viper.AutomaticEnv() // read in environment variables that match
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Error reading config file: ", err)
	}
}
