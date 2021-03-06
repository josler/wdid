package main

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"

	kingpin "github.com/alecthomas/kingpin"
	"github.com/josler/wdid/config"
	"github.com/josler/wdid/core"
	"github.com/josler/wdid/fileedit"
)

const (
	VERSION = "1.1.0"
)

var (
	app    = kingpin.New("wdid", "A tool to track what you did.")
	v      = app.Flag("verbose", "Enable verbose logging.").Short('v').Bool()
	format = app.Flag("format", "format to print in ('human', 'text', or 'json).").Default("human").Enum("human", "text", "json")

	bump     = app.Command("bump", "Bump an item to a new time, skipping the existing and creating a new one.")
	bumpID   = bump.Arg("id", "ID of item to bump.").Required().String()
	bumpTime = bump.Flag("time", "Time to bump item to the item at.").Short('t').PlaceHolder("TIME").Default("now").String()

	add      = app.Command("add", "Add a new task to track.")
	addTime  = add.Flag("time", "Time to add the task at.").Short('t').PlaceHolder("TIME").Default("now").String()
	addDone  = add.Flag("done", "Mark task as done already").Short('d').Bool()
	newThing = add.Arg("new-task", "Description of new task.").String()

	addNote      = app.Command("note", "Add a new note to track.")
	addNoteTime  = addNote.Flag("time", "Time to add the note at.").Short('t').PlaceHolder("TIME").Default("now").String()
	newNoteThing = addNote.Arg("new-note", "Summary of new note.").String()

	do   = app.Command("do", "Mark a task as done.")
	doID = do.Arg("id", "ID of task to mark done.").Required().String()

	edit            = app.Command("edit", "Edit an item's time or description.")
	editTime        = edit.Flag("time", "Time to add the item at.").Short('t').PlaceHolder("TIME").String()
	editID          = edit.Arg("id", "ID of item to edit.").Required().String()
	editDescription = edit.Arg("description", "Text of new item.").String()

	group        = app.Command("group", "create a group.")
	groupName    = group.Flag("name", "name of the group").Short('n').Required().String()
	groupFilters = group.Flag("filters", "filters for the group").Short('f').Required().String()

	groupRm     = app.Command("group-rm", "delete a group.")
	groupRmName = groupRm.Flag("name", "name of the group").Short('n').Required().String()

	groupList = app.Command("group-ls", "List groups.")

	importCmd      = app.Command("import", "Import items from a file or stdin.")
	importFilename = importCmd.Arg("in", "Filename to import from, if omitted, stdin used").String()

	list       = app.Command("ls", "List the items you're tracking.").Alias("list").Default()
	listFilter = list.Flag("filter", "Filter the results").Short('f').String()
	listGroup  = list.Flag("group", "List items in a group").Short('g').String()
	listArg    = list.Arg("filters", "Filter your items.").Default("0").String()

	rm   = app.Command("rm", "Remove (permanently!) a single item.").Alias("delete")
	rmID = rm.Arg("id", "ID of item to remove.").Required().String()

	skip   = app.Command("skip", "Mark a task as skipped.")
	skipID = skip.Arg("id", "ID of task to mark skipped.").Required().String()

	show          = app.Command("show", "Show a single item.")
	showID        = show.Arg("id", "ID of item to show.").Required().String()
	showConnected = show.Flag("connected", "Show connected items also.").Short('c').Bool()

	tagList = app.Command("tag-ls", "List tags.")
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
	ctx = context.WithValue(ctx, "verbose", *v)
	ctx = context.WithValue(ctx, "format", *format)
	ctx = context.WithValue(ctx, "config", conf)

	switch commandName {
	case add.FullCommand():
		var description io.Reader
		if *newThing != "" {
			description = strings.NewReader(*newThing)

		} else {
			description, err = fileedit.NewWithNoContent()
			if err != nil {
				break
			}
		}
		if *addDone {
			err = core.AddDone(ctx, description, *addTime)
		} else {
			err = core.Add(ctx, description, *addTime)
		}
	case addNote.FullCommand():
		var description io.Reader
		description, err = fileedit.EditExisting(*newNoteThing)
		if err != nil {
			break
		}
		err = core.AddNote(ctx, description, *addNoteTime)
	case bump.FullCommand():
		err = core.Bump(ctx, *bumpID, *bumpTime)
	case do.FullCommand():
		err = core.Do(ctx, *doID)
	case edit.FullCommand():
		if *editDescription == "" && *editTime == "" {
			err = core.EditDataFromFile(ctx, *editID)
		} else {
			err = core.Edit(ctx, *editID, strings.NewReader(*editDescription), *editTime)
		}
	case importCmd.FullCommand():
		err = core.Import(ctx, *importFilename)
	case list.FullCommand():
		if *listFilter != "" {
			*listArg = *listFilter // temporary override
		}
		err = core.List(ctx, *listArg, *listGroup)
	case rm.FullCommand():
		err = core.Rm(ctx, *rmID)
	case skip.FullCommand():
		err = core.Skip(ctx, *skipID)
	case show.FullCommand():
		err = core.Show(ctx, *showID, *showConnected)
	case tagList.FullCommand():
		err = core.ListTag(ctx)
	case group.FullCommand():
		err = core.CreateGroup(ctx, *groupName, *groupFilters)
	case groupRm.FullCommand():
		err = core.DeleteGroup(ctx, *groupRmName)
	case groupList.FullCommand():
		err = core.ListGroup(ctx)
	}
	app.FatalIfError(err, "")
}

func createStore(conf *config.Config) (core.Store, error) {
	switch conf.Store.Type {
	case "bolt":
		return core.NewBoltStore(conf.Store.Filepath())
	default:
		return nil, errors.New("store not specified correctly")
	}
}
