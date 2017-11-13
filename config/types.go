package config

type Datacenter struct {
	Name  string
	Keys  []Key
	Hosts []Host
}

type Host struct {
	Name string
	Port int
}

type Key struct {
	Key string
}
