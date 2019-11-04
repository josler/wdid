package main

import (
	"context"
	"errors"

	"github.com/asdine/storm"
	"github.com/josler/wdid/config"
	"github.com/josler/wdid/core"
)

func main() {
	conf, err := config.Load()
	if err != nil {
		return // TODO(JO): proper exit code
	}
	store, err := createStore(conf)
	if err != nil {
		return
	}
	defer store.Close()

	ctx := context.WithValue(context.Background(), "store", store)
	ctx = context.WithValue(ctx, "verbose", false)
	ctx = context.WithValue(ctx, "format", "human")
	ctx = context.WithValue(ctx, "config", conf)

	core.RenderUI(ctx)
}

func createStore(conf *config.Config) (core.Store, error) {
	switch conf.Store.Type {
	case "bolt":
		db, err := storm.Open(conf.Store.Filepath())
		if err != nil {
			return nil, err
		}
		return core.NewBoltStore(db), nil
	default:
		return nil, errors.New("store not specified correctly")
	}
}
