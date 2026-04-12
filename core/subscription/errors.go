package subscription

import "errors"

var (
	ErrUnsupportedFormat   = errors.New("unsupported subscription format")
	ErrInvalidSubscription = errors.New("invalid subscription")
	ErrNoProxyFound        = errors.New("no proxy found")
	ErrUnsupportedNodeType = errors.New("unsupported node type")
)
