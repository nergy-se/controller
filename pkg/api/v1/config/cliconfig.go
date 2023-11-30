package config

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"
)

type CliConfig struct {
	Server     string `default:"https://nergy.se"`
	APIToken   string
	TokenFile  string `default:"/etc/nergytoken"`
	SerialFile string `default:"/sys/firmware/devicetree/base/serial-number"`

	ControllerType string
	Address        string

	Serial string

	LogLevel string `default:"info"`

	mutex sync.RWMutex
}

func (c *CliConfig) Token() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.APIToken
}

func (c *CliConfig) SetToken(t string) {
	c.mutex.Lock()
	c.APIToken = strings.TrimSpace(t)
	c.mutex.Unlock()
}
func (c *CliConfig) PersistToken() error {
	if c.TokenFile == "" {
		return nil
	}
	return os.WriteFile(c.TokenFile, []byte(c.Token()), 0644)
}

func (c *CliConfig) LoadToken() error {
	if c.TokenFile == "" {
		return nil
	}
	if _, err := os.Stat(c.TokenFile); err == nil {
		b, err := os.ReadFile(c.TokenFile)
		if err != nil {
			return err
		}
		if len(b) == 0 {
			return nil // dont load empty token
		}

		c.SetToken(string(b))
	}
	return nil
}

func (c *CliConfig) SerialID() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.Serial
}

func (c *CliConfig) LoadSerial() error {
	id, err := os.ReadFile(c.SerialFile)
	if err != nil {
		return fmt.Errorf("error reading serialfile: %w", err)
	}
	c.mutex.Lock()
	c.Serial = string(bytes.TrimSpace(bytes.Trim(id, "\x00")))
	c.mutex.Unlock()
	return nil
}
