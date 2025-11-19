package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	S3 struct {
		Url            string `yaml:"url"`
		AccessKey      string `yaml:"accessKey"`
		SecretKey      string `yaml:"secretKey"`
		Region         string `yaml:"region"`
		MaxConnections int    `yaml:"maxConnections"`
		Bucket         string `yaml:"bucketName"`
	} `yaml:"s3"`
}

func GetConfig() (*Config, error) {
	yamlFile, err := os.ReadFile("config/config.yaml")
	if err != nil {
		return nil, err
	}
	var config Config

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
