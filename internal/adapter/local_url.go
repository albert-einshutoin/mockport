package adapter

import (
	"fmt"

	"github.com/albert-einshutoin/mockport/internal/config"
)

// LocalBaseURL returns the default local Mockport base URL for path.
func LocalBaseURL(path string) string {
	return fmt.Sprintf("http://localhost:%d%s", config.DefaultPort, path)
}
