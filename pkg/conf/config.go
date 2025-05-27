package conf

import (
	"github.com/kr/pretty"
	"path/filepath"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var (
	once sync.Once
	conf *Config
)

type Config struct {
	Server string `yaml:"server"`
	MySQL  MySQL  `yaml:"mysql"`
	Redis  Redis  `yaml:"redis"`
	MinIO  MinIO  `yaml:"minio"`
	JWT    JWT    `yaml:"jwt"`
}

type MySQL struct {
	DSN string `yaml:"dsn"`
}

type MinIO struct {
	Endpoint  string `yaml:"endpoint"`
	AccessKey string `yaml:"accessKey"`
	SecretKey string `yaml:"secretKey"`
}

type Redis struct {
	Addr string `yaml:"addr"`
}

type JWT struct {
	SignAlgo  string `yaml:"signAlgo"`
	SecretKey string `yaml:"secretKey"`
}

func GetConf() *Config {
	once.Do(func() {
		initConf()
	})
	return conf
}

func initConf() {
	prefix := "pkg/conf"
	path := filepath.Join(prefix, "conf.yml")

	env := filepath.Join(prefix, ".env")
	err := godotenv.Load(env)
	if err != nil {
		panic(err)
	}
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	conf = new(Config)
	if err := viper.Unmarshal(conf); err != nil {
		panic(err)
	}

	pretty.Printf("%#v\n", conf)
}
