package main

import (
	"fmt"
	"strings"
)

func greet(name string) string {
	if result := strings.TrimSpace(name); result != "" {
		return fmt.Sprintf("Hello, %s!", result)
	}

	return "Hello, World!"

}
