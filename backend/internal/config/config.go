package config

import (
	"strings"

	"github.com/spf13/viper"
)

// DBConfig 存放資料庫相關的設定
type DBConfig struct {
	Type string `mapstructure:"type"`
	Path string `mapstructure:"path"`
}

// Config 是整個應用程式的設定結構
type Config struct {
	Database DBConfig `mapstructure:"database"`
}

// LoadConfig 從設定檔或環境變數中讀取設定。
// path 參數指定設定檔所在的目錄。
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// 這允許我們透過環境變數來覆蓋設定，例如 DATABASE.PATH
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err = viper.ReadInConfig()
	if err != nil {
		// 如果只是設定檔不存在，可以忽略錯誤，因為可能完全依賴環境變數
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return
		}
	}

	err = viper.Unmarshal(&config)
	return
}
