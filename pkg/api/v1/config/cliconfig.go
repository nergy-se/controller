package config

import (
	"os"
	"sync"
)

type CliConfig struct {
	Server    string `default:"https://nergy.se"`
	APIToken  string
	TokenFile string `default:"/etc/nergytoken"`

	ControllerType string
	Address        string

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
	c.APIToken = t
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
