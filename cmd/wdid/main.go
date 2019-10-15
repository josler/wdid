package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	kingpin "github.com/alecthomas/kingpin"
	"github.com/asdine/storm"
	"github.com/josler/wdid/config"
	"github.com/josler/wdid/core"
	"github.com/josler/wdid/fileedit"
)

const (
	VERSION = "0.99.0"
)

var (
	app    = kingpin.New("wdid", "A tool to track what you did.")
	v      = app.Flag("verbose", "Enable verbose logging.").Short('v').Bool()
	format = app.Flag("format", "format to print in ('human', 'text', or 'json).").Default("human").Enum("human", "text", "json")

	auto     = app.Command("auto", "bring up items for automatic suggestion")
	autoTime = auto.Flag("time", "Time range to search in.").Short('t').PlaceHolder("TIME").Default("0").String()

	bump     = app.Command("bump", "Bump an item to a new time, skipping the existing and creating a new one.")
	bumpID   = bump.Arg("id", "ID of item to bump.").Required().String()
	bumpTime = bump.Flag("time", "Time to bump item to the item at.").Short('t').PlaceHolder("TIME").Default("now").String()

	add      = app.Command("add", "Add a new item to track.")
	addTime  = add.Flag("time", "Time to add the item at.").Short('t').PlaceHolder("TIME").Default("now").String()
	addDone  = add.Flag("done", "Mark item as done already").Short('d').Bool()
	newThing = add.Arg("new-item", "Description of new item.").String()

	do   = app.Command("do", "Mark an item as done.")
	doID = do.Arg("id", "ID of item to mark done.").Required().String()

	edit            = app.Command("edit", "Edit an item's time or description.")
	editTime        = edit.Flag("time", "Time to add the item at.").Short('t').PlaceHolder("TIME").String()
	editID          = edit.Arg("id", "ID of item to edit.").Required().String()
	editDescription = edit.Arg("description", "Description of new item.").String()

	group        = app.Command("group", "create a group.")
	groupName    = group.Flag("name", "name of the group").Short('n').Required().String()
	groupFilters = group.Flag("filters", "filters for the group").Short('f').Required().String()

	groupRm     = app.Command("group-rm", "delete a group.")
	groupRmName = groupRm.Flag("name", "name of the group").Short('n').Required().String()

	groupList = app.Command("group-ls", "List groups.")

	importCmd      = app.Command("import", "Import items from a file or stdin.")
	importFilename = importCmd.Arg("in", "Filename to import from, if omitted, stdin used").String()

	list         = app.Command("ls", "List the items you're tracking.").Alias("list").Default()
	listDone     = list.Flag("done", "Only list items with status = done.").Short('d').Bool()
	listWaiting  = list.Flag("waiting", "Only list items with status = waiting.").Short('w').Bool()
	listSkipped  = list.Flag("skipped", "Only list items with status = skipped.").Short('s').Bool()
	listBumped   = list.Flag("bumped", "Only list items with status = bumped.").Short('b').Bool()
	listFilter   = list.Flag("filter", "Filter the results").Short('f').String()
	listGroup    = list.Flag("group", "List items in a group").Short('g').String()
	listTime     = list.Arg("time", "Time range to search in.").Default("0").String()
	listTimeFlag = list.Flag("time", "Time range to search in.").Short('t').String()

	rm   = app.Command("rm", "Remove (permanently!) a single item.").Alias("delete")
	rmID = rm.Arg("id", "ID of item to remove.").Required().String()

	skip   = app.Command("skip", "Mark an item as skipped.")
	skipID = skip.Arg("id", "ID of item to mark skipped.").Required().String()

	show   = app.Command("show", "Show a single item.")
	showID = show.Arg("id", "ID of item to show.").Required().String()

	tag     = app.Command("tag", "work with tags.")
	tagList = app.Command("tag-ls", "List tags.")

	ui = app.Command("ui", "show a UI.")
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
	defer store.Close()

	ctx := context.WithValue(context.Background(), "store", store)
	ctx = context.WithValue(ctx, "verbose", *v)
	ctx = context.WithValue(ctx, "format", *format)
	ctx = context.WithValue(ctx, "config", conf)

	switch commandName {
	case add.FullCommand():
		var description io.Reader
		if *newThing != "" {
			description = strings.NewReader(*newThing)

		} else {
			description = os.Stdin
		}
		if *addDone {
			err = core.AddDone(ctx, description, *addTime)
		} else {
			err = core.Add(ctx, description, *addTime)
		}
	case auto.FullCommand():
		var confs []core.AutoConf
		for _, c := range conf.Auto { // dance around iface mapping
			confs = append(confs, c)
		}
		err = core.Auto(ctx, *autoTime, confs...)
	case bump.FullCommand():
		err = core.Bump(ctx, *bumpID, *bumpTime)
	case do.FullCommand():
		err = core.Do(ctx, *doID)
	case edit.FullCommand():
		if *editDescription == "" && *editTime == "" {
			err = editFromFile(ctx)
		} else {
			err = core.Edit(ctx, *editID, strings.NewReader(*editDescription), *editTime)
		}
	case importCmd.FullCommand():
		err = core.Import(ctx, *importFilename)
	case list.FullCommand():
		statuses := []string{}
		if *listBumped {
			statuses = append(statuses, core.BumpedStatus)
		}
		if *listDone {
			statuses = append(statuses, core.DoneStatus)
		}
		if *listWaiting {
			statuses = append(statuses, core.WaitingStatus)
		}
		if *listSkipped {
			statuses = append(statuses, core.SkippedStatus)
		}
		if *listTimeFlag != "" {
			err = core.List(ctx, *listTimeFlag, *listFilter, *listGroup, statuses...)
		} else {
			err = core.List(ctx, *listTime, *listFilter, *listGroup, statuses...)
		}
	case rm.FullCommand():
		err = core.Rm(ctx, *rmID)
	case skip.FullCommand():
		err = core.Skip(ctx, *skipID)
	case show.FullCommand():
		err = core.Show(ctx, *showID)
	case tagList.FullCommand():
		err = core.ListTag(ctx)
	case group.FullCommand():
		err = core.CreateGroup(ctx, *groupName, *groupFilters)
	case groupRm.FullCommand():
		err = core.DeleteGroup(ctx, *groupRmName)
	case groupList.FullCommand():
		err = core.ListGroup(ctx)
	case ui.FullCommand():
		fmt.Println("hi")
	}
	app.FatalIfError(err, "")
}

func editFromFile(ctx context.Context) error {
	fpath := config.ConfigDir() + "/WDID_TEMP"

	// find the item in question
	items, err := core.FindAll(ctx, *editID)
	if err != nil {
		return err
	}
	if len(items) != 1 {
		return errors.New("found too many items to edit")
	}

	data, err := fileedit.EditWithExistingContent(fpath, strings.NewReader(items[0].Data()))
	if err != nil {
		return err
	}
	return core.Edit(ctx, *editID, strings.NewReader(data), *editTime)
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
