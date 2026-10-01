package main

import (
	"fmt"
	"strings"
)

func greet(name string) string {
	name = strings.TrimSpace(name)
	if len(name) > 0 {
		return fmt.Sprintf("Hello, %s!", strings.TrimSpace(name))
	}
	return "Hello, World!"
}
