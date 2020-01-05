package config

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"main/config"
	"main/pubsub"
	"main/pubsub/systemevent"
	"time"
)

type ConfigManager struct {
	db     *gorm.DB
	config config.Config
}

func New(db *gorm.DB) *ConfigManager {
	db.AutoMigrate()
	return &ConfigManager{
		db: db,
	}
}

func (cm *ConfigManager) SetConfig(conf config.Config) error {
	// TODO: Validate
	cm.config = conf
	pubsub.UpdateConfigEvent.Pub(pubsub.UpdateConfig{cm.config})
	return nil
}

func (cm *ConfigManager) SetUpFromFile() error {
	body, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return err
	}

	conf := config.Config{}
	err = yaml.Unmarshal(body, &conf)
	if err != nil {
		return err
	}

	cm.SetConfig(conf)
	return nil
}

func (cm *ConfigManager) Get() config.Config {
	return cm.config
}

func (cm *ConfigManager) Save() error {
	// TODO: デフォルトの値が書き込まれてしまう(falseとか)
	buf, err := yaml.Marshal(cm.config)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile("config.yaml", buf, 0755); err != nil {
		return err
	}

	pubsub.SystemEvent.Pub(pubsub.System{Time: time.Now(), Type: systemevent.NEW_CONFIG_SAVE})
	return nil
}
