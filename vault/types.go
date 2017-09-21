package vault

// A struct to return the current status of the Vault server
type Status struct {
	Ready  bool
	Reason string
}
