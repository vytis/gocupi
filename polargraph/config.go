package polargraph

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/ilyakaznacheev/cleanenv"
)

type ConfigData struct {
	Board struct {
		Width float64 `yaml:"width"`
	} `yaml:"board"`
	Hardware struct {
		SpoolCircumference float64 `yaml:"spool_circumference_mm"`
		SinglStep          float64 `yaml:"single_step_degrees"`
		MaxAcceleration    float64 `yaml:"max_acceleration_s"`
		SerialPort         string  `yaml:"serial_port"`
	} `yaml:"hardware"`

	Position struct {
		left  float64 `yaml:"left" env-default:"0"`
		right float64 `yaml:"right" env-default:"0"`
	} `yaml:"position"`
}

// Mocking config reading
type ConfigInterface interface {
	ReadConfig(path string, cfg interface{}) error
	ConfigPath() string
	DefaultConfigPath() string
}

type ConfigReader struct{}

func (reader ConfigReader) ReadConfig(path string, cfg interface{}) error {
	return cleanenv.ReadConfig(path, cfg)
}

func (reader ConfigReader) ConfigPath() string {
	usr, _ := user.Current()
	return filepath.Join(usr.HomeDir, ".polargraph", settingsFile)
}

func (reader ConfigReader) DefaultConfigPath() string {
	var settingsFile string = "config.yml"
	_, filename, _, _ := runtime.Caller(0)
	repoBasepath := filepath.Dir(filepath.Dir(filename))
	return filepath.Join(repoBasepath, settingsFile)
}

var Config ConfigData

// Config reading implementation

func (config *ConfigData) Read() {
	config.read(ConfigReader{})
}

func (config *ConfigData) read(reader ConfigInterface) {
	configPath := reader.ConfigPath()

	err := reader.ReadConfig(configPath, config)
	if err != nil {
		configBasepath := filepath.Dir(configPath)
		repoSettingsFile := reader.DefaultConfigPath()

		if mkdirErr := os.MkdirAll(configBasepath, os.ModePerm); mkdirErr != nil {
			panic(mkdirErr)
		}
		if copyErr := copyFile(repoSettingsFile, configPath); copyErr != nil {
			panic(copyErr)
		}
		if err := reader.ReadConfig(repoSettingsFile, config); err != nil {
			panic(fmt.Sprint("Failed reading config file ", configPath, " :", err))
		}
	}
}
