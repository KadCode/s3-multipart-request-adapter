package config

import (
	"os"
	"time"

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
	FiberConfig struct {
		Prefork       bool          `yaml:"prefork"`
		CaseSensitive bool          `yaml:"case_sensitive"`
		StrictRouting bool          `yaml:"strict_routing"`
		ServerHeader  string        `yaml:"server_header"`
		AppName       string        `yaml:"app_name"`
		ReadTimeout   time.Duration `yaml:"read_timeout"`
		BodyLimit     int           `yaml:"body_limit"`
		Port          string        `yaml:"port"`
	} `yaml:"fiber"`
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

type TestConfig struct {
	S3 struct {
		Url               string `yaml:"url"`
		AccessKey         string `yaml:"accessKey"`
		SecretKey         string `yaml:"secretKey"`
		Session           string `yaml:"session"`
		Region            string `yaml:"region"`
		HostnameImmutable bool   `yaml:"hostnameImmutable"`
		MaxConnections    int    `yaml:"maxConnections"`
		Bucket            string `yaml:"bucketName"`
	} `yaml:"s3"`
}

func GetTestConfig() (*TestConfig, error) {
	yamlFile, err := os.ReadFile("../tests/test_config.yaml")
	if err != nil {
		return nil, err
	}
	var testConfig TestConfig

	err = yaml.Unmarshal(yamlFile, &testConfig)
	if err != nil {
		return nil, err
	}

	return &testConfig, nil
}
