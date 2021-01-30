package polargraph

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vytis/gocupi/polargraph/mocks"
)

func TestConfigNotFound(t *testing.T) {
	defaultConfig, _ := ioutil.TempFile("", "default_config.*.yml")
	defer os.Remove(defaultConfig.Name())
	config, _ := ioutil.TempFile("", "config.*.yml")
	defer os.Remove(config.Name())

	reader := new(mocks.ConfigInterface)
	var data ConfigData
	reader.On("ConfigPath").Return(config.Name())
	reader.On("ReadConfig", config.Name(), mock.Anything).Return(errors.New("error")).Once()
	reader.On("DefaultConfigPath").Return(defaultConfig.Name())
	reader.On("ReadConfig", defaultConfig.Name(), mock.Anything).Return(nil).Once()

	data.read(reader)

	reader.AssertExpectations(t)
}

func TestConfigFound(t *testing.T) {
	defaultConfig, _ := ioutil.TempFile("", "default_config.*.yml")
	defer os.Remove(defaultConfig.Name())
	config, _ := ioutil.TempFile("", "config.*.yml")
	defer os.Remove(config.Name())

	reader := new(mocks.ConfigInterface)
	var data ConfigData
	reader.On("ConfigPath").Return(config.Name())
	reader.On("ReadConfig", config.Name(), mock.Anything).Return(nil)

	data.read(reader)

	reader.AssertNotCalled(t, "DefaultConfigPath")
	reader.AssertExpectations(t)
}

func TestMoveConfigToUserDir(t *testing.T) {
	defaultConfig, _ := ioutil.TempFile("", "default_config.*.yml")
	defer os.Remove(defaultConfig.Name())
	config, _ := ioutil.TempFile("", "config.*.yml")
	os.Remove(config.Name())
	defer os.Remove(config.Name())

	testConfig := []byte("test config")
	if err := ioutil.WriteFile(defaultConfig.Name(), testConfig, 0644); err != nil {
		t.Error("Cannot write to config file")
	}

	reader := new(mocks.ConfigInterface)
	var data ConfigData
	reader.On("ConfigPath").Return(config.Name())
	reader.On("DefaultConfigPath").Return(defaultConfig.Name())
	reader.On("ReadConfig", config.Name(), mock.Anything).Return(errors.New("error")).Once()
	reader.On("ReadConfig", defaultConfig.Name(), mock.Anything).Return(nil).Once()

	data.read(reader)

	assert.FileExists(t, config.Name())

	dat, err := ioutil.ReadFile(config.Name())
	if err != nil {
		t.Error("Cannot read config file")
	}

	assert.ElementsMatch(t, testConfig, dat)
}
