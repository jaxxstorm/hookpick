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
	Port int
}

// Key struct
type Key struct {
	Key string
}
