package config

type Config struct {
	Module  string  `yaml:"module"`
	Service Service `yaml:"service"`
}

type Service struct {
	Name     string    `yaml:"name"`
	Logicals []Logical `yaml:"logicals"`
}

type Logical struct {
	Name         string   `yaml:"name"`
	API          API      `yaml:"api"`
	Dependencies []string `yaml:"dependencies"`
}

type API struct {
	FileName        string            `yaml:"fileName"`
	InterfaceName   string            `yaml:"interface"`
	Implementations APIImplementation `yaml:"implementations"`
}

type APIImplementation struct {
	Local bool `yaml:"local"`
	HTTP  bool `yaml:"http"`
}
