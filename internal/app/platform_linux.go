package app

import "context"

// claudeKeychain reports no Keychain login on Linux, where Claude Code uses .credentials.json.
func claudeKeychain(context.Context, string) ([]byte, error) {
	return nil, errSignedOut
}
