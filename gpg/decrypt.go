package gpg

import (
	"bytes"
	"encoding/base64"
	"os/exec"
)

func Decrypt(key string) (string, error) {

	var cmd exec.Cmd
	var output bytes.Buffer

	gpgCmd, err := exec.LookPath("gpg")

	if err != nil {
		return "", err
	}

	cmd.Path = gpgCmd
	cmd.Args = []string{"--decrypt", "--quiet"}

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
