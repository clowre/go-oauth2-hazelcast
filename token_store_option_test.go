package hcstore

import (
	"context"
	"testing"

	"github.com/hazelcast/hazelcast-go-client"
)

func TestHCTokenStoreOptions(t *testing.T) {

	type test struct {
		name      string
		options   []TokenStoreOption
		willError bool
	}

	table := []test{
		{
			name:      "empty map names",
			options:   []TokenStoreOption{WithAccessMapName(""), WithCodesMapName(""), WithRefreshMapName("")},
			willError: true,
		},
		{
			name:      "some empty names",
			options:   []TokenStoreOption{WithAccessMapName("am"), WithCodesMapName(""), WithRefreshMapName("rm")},
			willError: true,
		},
		{
			name:      "filled map names",
			options:   []TokenStoreOption{WithAccessMapName("am"), WithCodesMapName("cm"), WithRefreshMapName("rm")},
			willError: false,
		},
	}

	ctx := context.Background()
	hzClient, err := hazelcast.StartNewClient(ctx)
	if err != nil {
		t.Fatalf("failed to connect to hazelcast: %v", err)
	}
	defer hzClient.Shutdown(ctx)

	for _, x := range table {

		if _, err := NewTokenStore(hzClient, x.options...); err != nil {
			if !x.willError {
				t.Fatalf("expected to err on test %s", x.name)
			}
		} else {
			if x.willError {
				t.Fatalf("expected not to err on test %s", x.name)
			}
		}
	}
}
