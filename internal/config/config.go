package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	FacilityCollectionIDs map[string]string `yaml:"FacilityCollectionIDs"`
	GlobusScopes          []string          `yaml:"GlobusScopes"`
}

func ReadConfig() (Config, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return Config{}, err
	}
	executablePath, err := os.Executable()
	if err != nil {
		return Config{}, err
	}

	v := viper.New()
	v.AddConfigPath(filepath.Dir(executablePath))
	v.AddConfigPath(filepath.Join(userConfigDir, "globus-transfer-service"))
	v.SetConfigType("yaml")

	var conf Config
	err = v.UnmarshalExact(&conf)
	return conf, err
}
