package app

import (
	"context"
	"fmt"
)

func claudeKeychain(context.Context, string) ([]byte, error) {
	return nil, fmt.Errorf("credentials for Claude must be stored in .credentials.json on Linux")
}
