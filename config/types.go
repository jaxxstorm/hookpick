package config

type Datacenter struct {
	Name  string
	Key   string
	Hosts []Host
}

type Host struct {
	Name string
	Port int
}
