package main

import (
	"context"
	"errors"
	"os"

	"github.com/josler/wdid/migrations"

	"github.com/alecthomas/kingpin"
	"github.com/josler/wdid/config"
	"github.com/josler/wdid/core"
)

const VERSION = "0.1"

var (
	app         = kingpin.New("wdid_migrate", "migrations for wdid")
	addKinds    = app.Command("add_kinds", "Add kind to items. Items tagged #note should be Notes")
	addKindsArg = addKinds.Arg("from", "When should migration apply from?").Default("9000").String()
)

func main() {
	conf, err := config.Load()
	app.FatalIfError(err, "")
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.EnableFileExpansion = false

	app.Version(VERSION)
	app.HelpFlag.Short('h') // allow -h for --help
	app.UsageTemplate(kingpin.CompactUsageTemplate)
	app.Interspersed(true)

	commandName := kingpin.MustParse(app.Parse(os.Args[1:]))

	store, err := createStore(conf)
	app.FatalIfError(err, "")

	ctx := context.WithValue(context.Background(), "store", store)
	ctx = context.WithValue(ctx, "config", conf)

	switch commandName {
	case addKinds.FullCommand():
		migrations.AddKinds(ctx, *addKindsArg)
	}
}

func createStore(conf *config.Config) (core.Store, error) {
	switch conf.Store.Type {
	case "bolt":
		return core.NewBoltStore(conf.Store.Filepath())
	default:
		return nil, errors.New("store not specified correctly")
	}
}
