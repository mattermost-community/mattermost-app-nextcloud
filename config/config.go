package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	APPURL       string `mapstructure:"APPURL"`
	Port         string `mapstructure:"PORT"`
	STATICFOLDER string `mapstructure:"STATICFOLDER"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()

	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
