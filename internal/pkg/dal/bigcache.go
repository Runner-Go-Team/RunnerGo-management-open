package dal

import (
	"fmt"
	"time"

	"github.com/allegro/bigcache/v3"
)

var (
	bc *bigcache.BigCache
)

func MustInitBigCache() {
	var err error
	bc, err = bigcache.NewBigCache(bigcache.DefaultConfig(10 * time.Second))
	if err != nil {
		panic(fmt.Errorf("bigcache err %w", err))
	}
}

func BigCache() *bigcache.BigCache {
	return bc
}
