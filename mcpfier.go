package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

// Command represents the structure of a command in the config
type Command struct {
	Name   string   `yaml:"name"`
	Script string   `yaml:"script"`
	Args   []string `yaml:"args"`
}

// Config holds all the commands from the YAML file
type Config struct {
	Commands []Command `yaml:"commands"`
}

// LoadConfig loads the configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// RunCommand executes the command based on the configuration
func RunCommand(command Command) error {
	cmd := exec.Command(command.Script, command.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Command name required")
	}
	commandName := os.Args[1]

	config, err := LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	var foundCmd *Command
	for _, cmd := range config.Commands {
		if cmd.Name == commandName {
			foundCmd = &cmd
			break
		}
	}

	if foundCmd == nil {
		log.Fatalf("Command '%s' not found in config", commandName)
	}

	if err := RunCommand(*foundCmd); err != nil {
		log.Fatalf("Failed to run command '%s': %v", commandName, err)
	}
}
