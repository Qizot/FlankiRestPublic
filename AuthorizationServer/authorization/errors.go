package authorization

import "errors"

var (
	ErrTokenNotFound       = errors.New("Invalid access token")
	ErrDatabaseError       = errors.New("Database error while trying to fetch token")
	ErrUnauthorizedAccount = errors.New("Unauthorized account")
	ErrCrypto              = errors.New("bcrypt error")
)
