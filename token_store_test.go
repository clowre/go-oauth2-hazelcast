package hcstore

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/hazelcast/hazelcast-go-client"
)

func TestConnection(t *testing.T) {

	t.Log("testing with nil hazelcast client")
	if _, err := NewTokenStore(nil); err == nil {
		t.Fatal("expected to return an error on nil hc client")
	}

	t.Log("connecting to hc...")

	ctx := context.Background()
	hzClient, err := hazelcast.StartNewClient(ctx)
	if err != nil {
		t.Fatalf("failed to connect to hazelcast: %v", err)
	}

	t.Log("testing with ok hazelcast client...")
	store, err := NewTokenStore(hzClient)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("disconnecting from hc")
	if err := hzClient.Shutdown(ctx); err != nil {
		t.Fatalf("hc client shutdown failed: %v", err)
	}

	token := &models.Token{Access: "a", Code: "c", Refresh: "c"}
	if err := store.Create(ctx, token); err == nil {
		t.Fatal("expecting to return an error on closed hazelcast client")
	}

	t.Log("testing with disconnected hc client...")
	if _, err := NewTokenStore(hzClient); err == nil {
		t.Fatal("expected to return an error on a closed hazelcast client")
	}
}

func TestHazelcastTokenStore(t *testing.T) {

	ctx := context.Background()
	hzClient, err := hazelcast.StartNewClient(ctx)
	if err != nil {
		t.Fatalf("failed to connect to hazelcast: %v", err)
	}
	defer hzClient.Shutdown(ctx)

	store, err := NewTokenStore(
		hzClient,
		WithAccessMapName("access_tokens"),
		WithRefreshMapName("refresh_tokens"),
		WithCodesMapName("codes"),
	)
	if err != nil {
		t.Fatalf("cannot create hazelcast token store: %v", err)
	}

	testAccessStorage(ctx, t, store)
	testRefreshStorage(ctx, t, store)
	testCodeStorage(ctx, t, store)

	testStorageTTL(ctx, t, store)
}

func testAccessStorage(ctx context.Context, t *testing.T, store oauth2.TokenStore) {

	access := fmt.Sprintf("a_%d", time.Now().UnixMilli())
	t.Logf("testing access storage with token: %v", access)

	token := models.NewToken()
	token.SetAccess(access)
	token.SetAccessExpiresIn(5 * time.Minute)

	if err := store.Create(ctx, token); err != nil {
		t.Fatalf("cannot save token: %v", err)
	}

	newToken, err := store.GetByAccess(ctx, access)
	if err != nil {
		t.Fatalf("cannot find access token: %v", err)
	}

	if newToken.GetAccess() != access {
		t.Fatalf("retreived [%s] and actual [%s] access tokens did not match", newToken.GetAccess(), access)
	}

	if err := store.RemoveByAccess(ctx, access); err != nil {
		t.Fatalf("cannot remove access token [%s] from store: %v", access, err)
	}

	newToken, err = store.GetByAccess(ctx, access)
	if err == nil && newToken.GetAccess() == access {
		t.Fatalf("access token retreived from storage even after it was deleted")
	}

	t.Log("access storage tested")
}

func testRefreshStorage(ctx context.Context, t *testing.T, store oauth2.TokenStore) {

	refresh := fmt.Sprintf("r_%d", time.Now().UnixMilli())
	t.Logf("testing refresh storage with token: %v", refresh)

	token := models.NewToken()
	token.SetRefresh(refresh)
	token.SetRefreshExpiresIn(5 * time.Minute)

	if err := store.Create(ctx, token); err != nil {
		t.Fatalf("cannot save token: %v", err)
	}

	newToken, err := store.GetByRefresh(ctx, refresh)
	if err != nil {
		t.Fatalf("cannot find refresh token: %v", err)
	}

	if newToken.GetRefresh() != refresh {
		t.Fatalf("retreived [%s] and actual [%s] refresh tokens did not match", newToken.GetRefresh(), refresh)
	}

	if err := store.RemoveByRefresh(ctx, refresh); err != nil {
		t.Fatalf("cannot remove refresh token [%s] from store: %v", refresh, err)
	}

	newToken, err = store.GetByRefresh(ctx, refresh)
	if err == nil && newToken.GetRefresh() == refresh {
		t.Fatalf("refresh token retreived from storage even after it was deleted")
	}

	t.Log("refresh storage tested")
}

func testCodeStorage(ctx context.Context, t *testing.T, store oauth2.TokenStore) {

	code := fmt.Sprintf("c_%d", time.Now().UnixMilli())
	t.Logf("testing code storage with code: %v", code)

	token := models.NewToken()
	token.SetCode(code)
	token.SetCodeExpiresIn(5 * time.Minute)

	if err := store.Create(ctx, token); err != nil {
		t.Fatalf("cannot save token: %v", err)
	}

	newToken, err := store.GetByCode(ctx, code)
	if err != nil {
		t.Fatalf("cannot find code [%s]: %v", code, err)
	}

	if newToken.GetCode() != code {
		t.Fatalf("retreived [%s] and actual [%s] codes did not match", newToken.GetCode(), code)
	}

	if err := store.RemoveByCode(ctx, code); err != nil {
		t.Fatalf("cannot remove code [%s] from store: %v", code, err)
	}

	newToken, err = store.GetByCode(ctx, code)
	if err == nil && newToken.GetCode() == code {
		t.Fatalf("code retreived from storage even after it was deleted")
	}

	t.Log("code storage tested")
}

func testStorageTTL(ctx context.Context, t *testing.T, store oauth2.TokenStore) {

	t.Log("testing TTL")

	var (
		code    = fmt.Sprintf("code_%d", time.Now().UnixMilli())
		access  = fmt.Sprintf("access_%d", time.Now().UnixMilli())
		refresh = fmt.Sprintf("refresh_%d", time.Now().UnixMilli())
	)

	token := &models.Token{
		Code:             code,
		CodeExpiresIn:    3 * time.Second,
		Access:           access,
		AccessExpiresIn:  6 * time.Second,
		Refresh:          refresh,
		RefreshExpiresIn: 9 * time.Second,
	}

	t.Log("saving tokens")
	if err := store.Create(ctx, token); err != nil {
		t.Fatalf("failed to store token: %v", err)
	}

	t.Log("testing auth code ttl")
	time.Sleep(4 * time.Second)
	if _, err := store.GetByCode(ctx, code); err == nil {
		t.Fatal("auth code should be expired by now")
	}

	t.Log("testing access token ttl")
	time.Sleep(4 * time.Second)
	if _, err := store.GetByAccess(ctx, access); err == nil {
		t.Fatal("access token should be expired by now")
	}

	t.Log("testing refresh token ttl")
	time.Sleep(4 * time.Second)
	if _, err := store.GetByRefresh(ctx, refresh); err == nil {
		t.Fatal("refresh token should be expired by now")
	}
}
