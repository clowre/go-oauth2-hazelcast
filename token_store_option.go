package hcstore

import "github.com/pkg/errors"

// TokenStoreOption is a function that can be used to modify behavior of the `tokenStore`.
type TokenStoreOption func(ts *tokenStore) error

// WithAccessMapName sets the name of the map that is used to save access tokens. An error is returned
// if the name is empty.
func WithAccessMapName(name string) TokenStoreOption {
	return func(ts *tokenStore) error {
		if name == "" {
			return errors.New("invalid access map name")
		}

		ts.accessMapName = name
		return nil
	}
}

// WithRefreshMapName sets the name of the map that is used to save refresh tokens. An error is returned
// if the name is empty.
func WithRefreshMapName(name string) TokenStoreOption {
	return func(ts *tokenStore) error {
		if name == "" {
			return errors.New("invalid refresh map name")
		}

		ts.refreshMapName = name
		return nil
	}
}

// WithCodesMapName sets the name of the map that is used to save codes. An error is returned if the name
// is empty.
func WithCodesMapName(name string) TokenStoreOption {
	return func(ts *tokenStore) error {
		if name == "" {
			return errors.New("invalid code map name")
		}

		ts.codeMapName = name
		return nil
	}
}
