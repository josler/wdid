package main

import (
	"context"
	"os"
	"strings"

	kingpin "github.com/alecthomas/kingpin"
	"github.com/josler/wdid/core"
)

const (
	VERSION = "0.99.0"
)

var (
	app    = kingpin.New("wdid", "A tool to track what you did.")
	v      = app.Flag("verbose", "Enable verbose logging.").Short('v').Bool()
	format = app.Flag("format", "format to print in ('human' or 'text').").Default("human").Enum("human", "text")

	auto     = app.Command("auto", "bring up items for automatic suggestion")
	autoTime = auto.Flag("time", "Time range to search in.").Short('t').PlaceHolder("TIME").Default("0").String()

	bump     = app.Command("bump", "Bump an item to a new time, skipping the existing and creating a new one.")
	bumpID   = bump.Arg("id", "ID of item to bump.").Required().String()
	bumpTime = bump.Flag("time", "Time to bump item to the item at.").Short('t').PlaceHolder("TIME").Default("now").String()

	add      = app.Command("add", "Add a new item to track.")
	addTime  = add.Flag("time", "Time to add the item at.").Short('t').PlaceHolder("TIME").Default("now").String()
	newThing = add.Arg("new-item", "Description of new item.").String()

	do   = app.Command("do", "Mark an item as done.")
	doID = do.Arg("id", "ID of item to mark done.").Required().String()

	edit            = app.Command("edit", "Edit an item's time or description.")
	editTime        = edit.Flag("time", "Time to add the item at.").Short('t').PlaceHolder("TIME").String()
	editID          = edit.Arg("id", "ID of item to edit.").Required().String()
	editDescription = edit.Arg("description", "Description of new item.").String()

	group        = app.Command("group", "create a group.")
	groupName    = group.Arg("name", "name of the group").Required().String()
	groupFilters = group.Flag("filters", "filters for the group").Short('f').Required().String()

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

	rm   = app.Command("rm", "Remove (permanently!) a single item.")
	rmID = rm.Arg("id", "ID of item to remove.").Required().String()

	skip   = app.Command("skip", "Mark an item as skipped.")
	skipID = skip.Arg("id", "ID of item to mark skipped.").Required().String()

	show   = app.Command("show", "Show a single item.")
	showID = show.Arg("id", "ID of item to show.").Required().String()

	tag     = app.Command("tag", "work with tags.")
	tagList = tag.Command("ls", "List tags.").Alias("list").Default()
)

func main() {
	conf, err := loadConfig()
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

	switch commandName {
	case add.FullCommand():
		if *newThing != "" {
			err = core.Add(ctx, strings.NewReader(*newThing), *addTime)
		} else {
			err = core.Add(ctx, os.Stdin, *addTime)
		}
	case auto.FullCommand():
		var confs []core.AutoConf
		for _, c := range conf.Auto { // dance around iface mapping
			confs = append(confs, c)
		}
		core.Auto(ctx, *autoTime, confs...)
	case bump.FullCommand():
		err = core.Bump(ctx, *bumpID, *bumpTime)
	case do.FullCommand():
		err = core.Do(ctx, *doID)
	case edit.FullCommand():
		err = core.Edit(ctx, *editID, strings.NewReader(*editDescription), *editTime)
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
	}
	app.FatalIfError(err, "")
}
