// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

type Config struct {
	Address                string            `config:"address"`
	Port                   int               `config:"port"`
	Timeout                uint32            `config:"timeout"`
	Meta                   bool              `config:"meta"`
	EnableSSL              bool              `config:"ssl.enable"`
	SSLCrt                 string            `config:"ssl.certificate"`
	SSLKey                 string            `config:"ssl.key"`
}

var DefaultConfig = Config{
	Address:                "127.0.0.1",
	Port:                   5044,
	Timeout:                0,
	Meta:			false,
	EnableSSL:              false,
	SSLCrt:			"",
	SSLKey:			"",
}
