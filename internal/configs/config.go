package configs

import (
	"fmt"

	"github.com/spf13/viper"
)


type Config struct {
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	JWTSecret  string `mapstructure:"JWT_SECRET"`
	JWTExpiry  int    `mapstructure:"JWT_EXPIRY"`
	GoogleClientID     string `mapstructure:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `mapstructure:"GOOGLE_CLIENT_SECRET"`
}

func LoadConf() (*Config, error)  {
  viper.SetConfigFile(".env")
  viper.SetConfigType("env")
  viper.AutomaticEnv()
  
  err := viper.ReadInConfig()
  if err != nil {
    if _, ok := err.(viper.ConfigFileNotFoundError); ok {
      return nil, nil
    }
    return nil, fmt.Errorf("gagal membaca config: %w", err)
  }
  
  var config Config
  err = viper.Unmarshal(&config)
  if err != nil {
    return nil, fmt.Errorf("gagal unmarshal config: %w", err)
  }
  
  return &config, nil
}
