package gpg

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"os/exec"
)

type StringDecrypter func(string) (string, error)

type GPGHelper struct {
	Decrypt StringDecrypter
}

func NewGPGHelper(decrypter StringDecrypter) *GPGHelper {
	return &GPGHelper{
		Decrypt: decrypter,
	}
}

func gpg_major_version() (int, error) {
	cmdName := "gpg"
	cmdArgs := []string{"--version"}

	var cmdOut []byte
	var err error

	if cmdOut, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		fmt.Println("There was an error running gpg --version: ", err)
		return -1, err
	}
	gpgvers := string(cmdOut)

	// gpg1 --version output should look similar to
	//
	//	gpg (GnuPG) 1.4.20
	//	Copyright (C) 2015 Free Software Foundation, Inc.
	//...

	// gpg2 --version
	//	gpg (GnuPG) 2.2.5
	//	libgcrypt 1.8.2
	//	Copyright (C) 2018 Free Software Foundation, Inc.

	for i := 0; i < len(gpgvers); i++ {
		if gpgvers[i] == '1' {
			return 1, nil
		}
		if gpgvers[i] == '2' {
			return 2, nil
		}
	}
	return -1, errors.New("Could not determine gpg major version")
}

// Decrypt GPG keys
func Decrypt(key string) (string, error) {

	var cmd exec.Cmd
	var output bytes.Buffer
	var gpgvers int

	gpgCmd, err := exec.LookPath("gpg")

	if err != nil {
		return "", err
	}

	gpgvers, err = gpg_major_version()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Due to error determining gpg version, defaulting to gpg vers 1 options")
		gpgvers = 1
	}

	cmd.Path = gpgCmd
	if gpgvers == 1 {
		fmt.Println("Using GPG decrypt for GPG version 1")
		cmd.Args = []string{"--decrypt", "--quiet"}
	}
	if gpgvers == 2 {
		fmt.Println("Using GPG decrypt for GPG version 2")
		cmd.Args = []string{"--decrypt", "--quiet", "--pinentry-mode", "loopback"}
	}
	dec, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return "", err
	}

	// return the reader interface for dec (byte array)
	d := bytes.NewReader(dec)

	// pipe d to gpg commands stdin
	cmd.Stdin = d
	cmd.Stdout = &output

	if err := cmd.Run(); err != nil {
		return "", err
	}

	// return the output from the gpg command
	return output.String(), nil

}
