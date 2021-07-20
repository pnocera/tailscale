package main

import (
	"github.com/spf13/viper"
)

//Config struct using viper
type Config struct {
	v *viper.Viper
}

//New Create a new config
func NewConfig() *Config {
	c := Config{
		v: viper.New(),
	}

	c.v.SetEnvPrefix("")
	c.v.AutomaticEnv()

	return &c
}

func (c *Config) ForwardHost() string {
	return c.v.GetString("FORWARD_HOST")
}

func (c *Config) ListenHostPort() string {
	return c.v.GetString("LISTEN_HOSTPORT")
}
