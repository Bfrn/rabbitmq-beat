package sse

import (
	"fmt"
	"strconv"
)

type config struct {
	Port string `config:"port"`
}

var (
	defaultConfig = config{
		Port: "8080",
	}
)

func (c *config) Validate() error {
	i, err := strconv.Atoi(c.Port)

	if err != nil {
		return fmt.Errorf("The specified port should be a number")
	}

	if i < 1 {
		return fmt.Errorf("The specified port must be greate than 0")
	}

	return nil
}
