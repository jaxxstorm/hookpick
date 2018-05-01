package config

// Datacenter struct
type Datacenter struct {
	Name  string
	Keys  []Key
	Hosts []Host
}

// Host struct
type Host struct {
	Name string
	Port string
}

// Key struct
type Key struct {
	Key string
}
