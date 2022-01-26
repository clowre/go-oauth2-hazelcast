# Hazelcast store for [go-oauth2/oauth2](https://github.com/go-oauth2/oauth2)

The store requires a runnig `*hazelcast.Client` to manage tokens and codes.

```go
package main 

import (
    "context"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/hazelcast/hazelcast-go-client"

    "github.com/clowre/go-oauth2-hazelcast"
)

func main() {
    
    ctx := context.Background()
    client, err := hazelcast.StartNewClient(ctx)
    if err != nil {
        panic(err)
    }
    defer client.Shutdown()

    store, err := hcstore.NewTokenStore(
        client,
		WithAccessMapName("access_tokens"),
		WithRefreshMapName("refresh_tokens"),
		WithCodesMapName("codes"),
    )
    if err != nil {
        panic(err)
    }
}
```

The tests for this package assumes that there is a Hazelcast cluster running on `localhost:5701`.