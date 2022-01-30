package hcstore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/pkg/errors"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/hazelcast/hazelcast-go-client"
)

// NewTokenStore creates an instances of `oauth2.TokenStore` connected to a Hazelcast cluster. This implementation
// relies on Hazelcast maps to manage access tokens, refresh tokens, and codes. By default, access tokens, refresh
// toknes, and codes are stored in maps oauth2_access_tokens, oauth2_refresh_tokens, and oauth2_codes respectively,
// but this can be changed by `WithAccessMapName`, `WithRefreshMapName`, and `WithCodesMapName` options.
// This package assumes that it will be supplied with a valid Hazelcast client, leaving the connecting/disconnecting
// to the cluser to its users.
func NewTokenStore(client *hazelcast.Client, opts ...TokenStoreOption) (oauth2.TokenStore, error) {

	if client == nil {
		return nil, errors.New("hazelcast client must not be nil")
	}

	if !client.Running() {
		return nil, errors.New("hazelcast client is not running")
	}

	ts := &tokenStore{
		client:         client,
		accessMapName:  "oauth2_access_tokens",
		refreshMapName: "oauth2_refresh_tokens",
		codeMapName:    "oauth2_codes",

		// use a pool of `bytes.Buffer`s to avoid redeclaring byte slices when `Create` method is called.
		bufferPool: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}

	for _, o := range opts {
		if err := o(ts); err != nil {
			return nil, err
		}
	}

	return ts, nil
}

type tokenStore struct {
	client                                     *hazelcast.Client
	accessMapName, refreshMapName, codeMapName string
	bufferPool                                 sync.Pool
}

// create and store the new token information
func (h *tokenStore) Create(ctx context.Context, info oauth2.TokenInfo) error {

	buf := h.bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer h.bufferPool.Put(buf)

	if err := json.NewEncoder(buf).Encode(info); err != nil {
		return errors.Wrap(err, "cannot encode token to json")
	}

	ts := buf.String()

	if info.GetAccess() != "" {
		if err := h.putAccessToken(ctx, info, ts); err != nil {
			return err
		}
	}

	if info.GetRefresh() != "" {
		if err := h.putRefreshToken(ctx, info, ts); err != nil {
			return err
		}
	}

	if info.GetCode() != "" {
		if err := h.putCode(ctx, info, ts); err != nil {
			return err
		}
	}

	return nil
}

// delete the authorization code
func (h *tokenStore) RemoveByCode(ctx context.Context, code string) error {

	tm, err := h.client.GetMap(ctx, h.codeMapName)
	if err != nil {
		return err
	}

	if _, err := tm.Remove(ctx, fmt.Sprintf(`code:%s`, code)); err != nil {
		return err
	}

	return nil
}

// use the access token to delete the token information
func (h *tokenStore) RemoveByAccess(ctx context.Context, access string) error {

	tm, err := h.client.GetMap(ctx, h.accessMapName)
	if err != nil {
		return err
	}

	if _, err := tm.Remove(ctx, fmt.Sprintf(`access:%s`, access)); err != nil {
		return err
	}

	return nil
}

// use the refresh token to delete the token information
func (h *tokenStore) RemoveByRefresh(ctx context.Context, refresh string) error {

	tm, err := h.client.GetMap(ctx, h.refreshMapName)
	if err != nil {
		return err
	}

	if _, err := tm.Remove(ctx, fmt.Sprintf(`refresh:%s`, refresh)); err != nil {
		return err
	}

	return nil
}

// use the authorization code for token information data
func (h *tokenStore) GetByCode(ctx context.Context, code string) (oauth2.TokenInfo, error) {

	tm, err := h.client.GetMap(ctx, h.codeMapName)
	if err != nil {
		return nil, err
	}

	val, err := tm.Get(ctx, fmt.Sprintf(`code:%s`, code))
	if err != nil {
		return nil, err
	}

	valString, ok := val.(string)
	if !ok {
		return nil, errors.New("retrieved code information was not a string")
	}

	token := models.NewToken()
	if err := json.Unmarshal([]byte(valString), token); err != nil {
		return nil, err
	}

	return token, nil
}

// use the access token for token information data
func (h *tokenStore) GetByAccess(ctx context.Context, access string) (oauth2.TokenInfo, error) {

	tm, err := h.client.GetMap(ctx, h.accessMapName)
	if err != nil {
		return nil, err
	}

	val, err := tm.Get(ctx, fmt.Sprintf(`access:%s`, access))
	if err != nil {
		return nil, err
	}

	valString, ok := val.(string)
	if !ok {
		return nil, errors.New("retrieved access information was not a string")
	}

	token := models.NewToken()
	if err := json.Unmarshal([]byte(valString), token); err != nil {
		return nil, err
	}

	return token, nil
}

// use the refresh token for token information data
func (h *tokenStore) GetByRefresh(ctx context.Context, refresh string) (oauth2.TokenInfo, error) {

	tm, err := h.client.GetMap(ctx, h.refreshMapName)
	if err != nil {
		return nil, err
	}

	val, err := tm.Get(ctx, fmt.Sprintf(`refresh:%s`, refresh))
	if err != nil {
		return nil, err
	}

	valString, ok := val.(string)
	if !ok {
		return nil, errors.New("retrieved refresh information was not a string")
	}

	token := models.NewToken()
	if err := json.Unmarshal([]byte(valString), token); err != nil {
		return nil, err
	}

	return token, nil
}

func (h *tokenStore) putAccessToken(ctx context.Context, info oauth2.TokenInfo, raw string) error {

	tm, err := h.client.GetMap(ctx, h.accessMapName)
	if err != nil {
		return err
	}

	_, err = tm.PutWithTTL(ctx, fmt.Sprintf(`access:%s`, info.GetAccess()), raw, info.GetAccessExpiresIn())
	if err != nil {
		return err
	}

	return nil
}

func (h *tokenStore) putRefreshToken(ctx context.Context, info oauth2.TokenInfo, raw string) error {

	tm, err := h.client.GetMap(ctx, h.refreshMapName)
	if err != nil {
		return err
	}

	_, err = tm.PutWithTTL(ctx, fmt.Sprintf(`refresh:%s`, info.GetRefresh()), raw, info.GetRefreshExpiresIn())
	if err != nil {
		return err
	}

	return nil
}

func (h *tokenStore) putCode(ctx context.Context, info oauth2.TokenInfo, raw string) error {

	tm, err := h.client.GetMap(ctx, h.codeMapName)
	if err != nil {
		return err
	}

	_, err = tm.PutWithTTL(ctx, fmt.Sprintf(`code:%s`, info.GetCode()), raw, info.GetCodeExpiresIn())
	if err != nil {
		return err
	}

	return nil
}

// API Check
var _ oauth2.TokenStore = (*tokenStore)(nil)
