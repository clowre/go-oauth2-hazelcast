package hcstore

import (
	"context"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/hazelcast/hazelcast-go-client"
)

func NewTokenStore(client *hazelcast.Client) oauth2.TokenStore {
	return &hzTokenStore{client: client}
}

type hzTokenStore struct {
	client *hazelcast.Client
}

// create and store the new token information
func (h *hzTokenStore) Create(ctx context.Context, info oauth2.TokenInfo) error {
	panic("not implemented") // TODO: Implement
}

// delete the authorization code
func (h *hzTokenStore) RemoveByCode(ctx context.Context, code string) error {
	panic("not implemented") // TODO: Implement
}

// use the access token to delete the token information
func (h *hzTokenStore) RemoveByAccess(ctx context.Context, access string) error {
	panic("not implemented") // TODO: Implement
}

// use the refresh token to delete the token information
func (h *hzTokenStore) RemoveByRefresh(ctx context.Context, refresh string) error {
	panic("not implemented") // TODO: Implement
}

// use the authorization code for token information data
func (h *hzTokenStore) GetByCode(ctx context.Context, code string) (oauth2.TokenInfo, error) {
	panic("not implemented") // TODO: Implement
}

// use the access token for token information data
func (h *hzTokenStore) GetByAccess(ctx context.Context, access string) (oauth2.TokenInfo, error) {
	panic("not implemented") // TODO: Implement
}

// use the refresh token for token information data
func (h *hzTokenStore) GetByRefresh(ctx context.Context, refresh string) (oauth2.TokenInfo, error) {
	panic("not implemented") // TODO: Implement
}
