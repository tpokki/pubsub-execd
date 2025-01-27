package main

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Logging  LogConfig       `yaml:"logging"`
	Triggers []TriggerConfig `yaml:"triggers"`
}

type LogConfig struct {
	Output string `yaml:"output"`
}

type TriggerConfig struct {
	Name   string       `yaml:"name"`
	PubSub PubSubConfig `yaml:"pubsub"`
	Run    RunConfig    `yaml:"run"`
}

type PubSubConfig struct {
	Project      string `yaml:"project"`
	Subscription string `yaml:"subscription"`
}

type RunConfig struct {
	Exec        string        `yaml:"exec"`
	Args        ArgsConfig    `yaml:"args"`
	Timeout     time.Duration `yaml:"timeout"`
	Concurrency int           `yaml:"concurrency"`
}

type ArgsConfig struct {
	Expression string `yaml:"expression"`
}

func loadConfig(file string) Config {
	var config Config
	// read file and unmarshal it
	data, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}
	return config
}

func (c *RunConfig) UnmarshalYAML(value *yaml.Node) error {
	var tmp struct {
		Exec        string     `yaml:"exec"`
		Args        ArgsConfig `yaml:"args"`
		Timeout     string     `yaml:"timeout"`
		Concurrency int        `yaml:"concurrency"`
	}
	err := value.Decode(&tmp)
	if err != nil {
		return err
	}

	timeout, err := time.ParseDuration(tmp.Timeout)
	if err != nil {
		return err
	}

	*c = RunConfig{
		Exec:        tmp.Exec,
		Args:        tmp.Args,
		Timeout:     timeout,
		Concurrency: tmp.Concurrency,
	}

	return nil
}
