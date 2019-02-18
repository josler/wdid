### Overview

What Did I Do (wdid) is a small CLI tool to track what you have been working on. You can `add`, `list`, `edit`, `do`, `skip`, `bump`, `show` and `rm` items. Alongside this manual tracking, there's an `auto` feature that can help automate fetching information about what you've done.

There's tagging built-in, and the ability to search your items by tag. Using this simple foundational item, you can organise your data however you like.

This tool both aims to track your most important goals, day-to-day, and help track what you have actually been working on in detail. Often when working we have goals. Goals are easy to track, there's a known outcome ahead of time. What's much harder is answering the question "where did all my time go?". Wdid aims to address that.

```
$ wdid help
usage: wdid [<flags>] <command> [<args> ...]

A tool to track what you did.

Flags:
  -h, --help          Show context-sensitive help (also try --help-long and --help-man).
  -v, --verbose       Enable verbose logging.
      --format=human  format to print in ('human' or 'text').
      --version       Show application version.

Commands:
  help [<command>...]
  auto [<flags>]
  bump [<flags>] <id>
  add [<flags>] [<new-item>]
  do <id>
  edit [<flags>] <id> [<description>]
  import [<in>]
  ls* [<flags>] [<time>]
  rm <id>
  skip <id>
  show <id>
  tag
    ls*
```

### Installation

```
go get -u github.com/josler/wdid/...
```


### Usage

#### add

```shell
$ wdid add "my task item"
$ wdid add -t 1 "my task from yesterday that I forgot."
```

You can also add from stdin:

```shell
$ wdid add < myitem.txt
```

#### show

Calling `show` with an ID shows more detail on the item.

```shell
$ wdid show a9fi3q
⇒ a9fi3q -- Mon, 26 Mar 2018 00:00:00
InternalID: recJyUxvMHSao4xZ9
Data:
 my task from yesterday that I forgot.
```

You can also just a prefix for the ID, and wdid will attempt to match the correct one - within a time frame of the last 14 days.

```shell
$ wdid show a9f
⇒ a9fi3q -- Mon, 26 Mar 2018 00:00:00
InternalID: recJyUxvMHSao4xZ9
Data:
 my task from yesterday that I forgot.
```

#### edit

You can edit the description or the time of an item. For example, to change the description and set the time to the start of today:

```shell
$ wdid edit a9fi3q "my new description" -t day
```

#### list

`ls` or `list` is the default subcommand, listing all tasks from a period of time (default today).

```shell
$ wdid
⇒ l72i3q  "my task item"                                    Tue, 27 Mar 2018 19:10:40
$ wdid ls # equivalent
⇒ l72i3q  "my task item"                                    Tue, 27 Mar 2018 19:10:40
```

You can also pass time structures to the list command.

```shell
$ wdid list week # all tasks from this week
⇒ a9fi3q  "my task from yesterday that I forgot."           Mon, 26 Mar 2018 00:00:00
⇒ l72i3q  "my task item"                                    Tue, 27 Mar 2018 19:10:40
```

You can list by item status too:

```shell
$ wdid ls -d # done tasks from this week
$ wdid ls -s # skipped tasks from this week
$ wdid ls -w # waiting tasks from this week
$ wdid ls -b # bumped tasks from this week
```

These can also be combined:

```shell
$ wdid list -sb month # skipped and bumped tasks from this month
```

There's also an advanced listing filter language, please see details below.

#### do

Items in wdid can be in one of four states:

- waiting: items to be worked on.
- skipped: items that have been skipped/dropped and no longer are waiting to be done.
- bumped: items that have been bumped forward (carried over) to another time.
- done: items that have been completed.

Items start in a waiting state, and then can be moved to done with `do`, and be marked with a green tick:

```shell
$ wdid do a9f
✔ a9fi3q -- Mon, 26 Mar 2018 00:00:00
InternalID: recJyUxvMHSao4xZ9
Data:
 my task from yesterday that I forgot.
```

#### skip

Items can be moved to skipped with `skip`, and be marked with a red x:

```shell
$ wdid skip a9f
✘ a9fi3q -- Mon, 26 Mar 2018 00:00:00
InternalID: recJyUxvMHSao4xZ9
Data:
 my task from yesterday that I forgot.
```

#### bump

Items can be bumped or carried forward with `bump`. This will return a new 'waiting' item, linked to the old one:

```shell
$ wdid bump a9f
⇒ i3nh99 -- Tue, 27 Mar 2018 19:20:44
InternalID: recjj9d4MH3QmI73t
Bumped from: a9fi3q
Data:
 my task from yesterday that I forgot.
```

The old item gets marked as bumped, have a reference to the new item, and be marked with a yellow ⇒:

```shell
$ wdid show a9f
⇒ a9fi3q -- Mon, 26 Mar 2018 00:00:00
InternalID: recJyUxvMHSao4xZ9
Bumped to: i3nh99
Data:
 my task from yesterday that I forgot.
```

Times can also be passed to the `bump` command to bump to a paricular time:

```shell
$ wdid bump yyt week # bump a task from the past to the start of the week.
```

#### rm

Items can also be hard deleted. Gone forever.

```shell
$ wdid rm i3nh99
```

#### tag list

Items can be tagged, and we can use the tag list command to show all tags we've created so far (not which items were tagged, but the tags themselves).

```
$ wdid tag list
@josler
#pr
#meeting
```

You can search for items by tag with our advanced listing filter language, please see below. More details of how tags work can also be found below.

### Viewing Data

Data can be printed in a couple of different ways. The two supported formats are "text" and "human". The text format is tab-delimited and useful for parsing with other command line tools, whereas the human format is easier to read for humans (colored, unicode characters, more detail when viewing single items). The default is "human". To change, pass a "format" flag: `wdid list --format=text week`.

The text format is especially helpful for exporting and importing data:

#### export

Data can be exported to text through the list command with text format. For example, to write the last 14 days worth of data to text, you can use the following:

```shell
wdid list --format=text 14 > file.txt
```

To view, `column` works nicely:

```shell
column -t -s $'\t' file.txt
```

#### import

Data can be imported in text format from a file or stdin.

```shell
wdid import file.txt
```

```shell
cat file.txt | wdid import
```

Imported items will overwrite duplicates of that item.

### Time parsing

Times can be passed in the following formats:

- `now`: Now until end of day.
- `0`: Start of today (midnight in your TZ) - equates to "today" when searching. Equivalent to `day`. Ends end of today.
- Integer n (e.g. `1`, `6`): start of the day, n days ago - equates to "in the last n days" when searching. Ends end of today.
- `day`: Start of today (midnight in your TZ). Equivalent to `1`. Ends end of today.
- `week`: Start of the week (monday, midnight in your TZ) - equates to "in the last week" when searching. Ends end of the week.
- `month`: Start of the month (first day of month, midnight in your TZ) - equates to "in the last month" when searching. Ends end of the month.
- `yesterday`: Yesterday.
- `today`: Today.
- `tomorrow`: Tomorrow.
- `monday`: Start Monday of _this week_. Ends end of that day. Same for every day of the week. Can also use short forms like `tue` or `tues`.
- `this monday`: Start Monday of _this week_. Ends that day.
- `next monday` Start Monday of _next week_. Ends that day.
- `last monday` Start Monday of _last week_. Ends that day.
- `last week` Start at start of previous week, end at end of that week.
- `next week` Start at start of next week, end at end of that week.
- `last month` Start at start of last month, end at end of that month.
- `next month` Start at start of next month, end at end of that month.
- `YYYY-MM-DD` (`2006-01-02` in Go time format): Start of given day in your TZ. Ends end of that day.
- `YYYY-MM-DDTHH:MM`: particular time on a day in your TZ. Ends end of that day.

Note that these times cover a _range_ of values. Usually from the start of the indicated day (00:00) to the end of the day (23:59) at the end of the period, inclusive.

When adding items, or setting the time for an item, wdid uses the _start_ of the period to do so. When searching for items, wdid uses the range. This sounds more complicated than it is, in practise it does what you'd expect.

### Tags

Wdid supports using tags to mark items, but the way it does this is not through manual tagging, but by parsing the item text itself. There are several ways to indicate through text that you'd like an item to have a tag.

```shell
$ wdid add "send invoice #project1 #billing" # two tags, "#project1" and "#billing".
$ wdid add "ask @josler to work more" # tag "@josler"
$ wdid add "[tag, another]" # two tags, "#tag" and "#another"
```

In general, any word with a preceeding '#' or '@' will be used as a tag, and anything within square brackets as well (anything without a leading '#' or '@' will have a '#' added). The leading '#' or '@' is part of the tag name.

You can then search for items with particular tags using our advanced listing.

### Advanced listing

There's a basic filter/query language built into wdid. You can use it to filter results in a more powerful way that the presets. To use it, pass the `--filter, -f` flag:

```shell
$ wdid list --filter "tag=#pr,status=waiting,time=week" # show me all of the items tagged "#pr", with a status of "waiting" from this week.

# we don't need the "list" command either, as it's default
$ wdid -f "tag=#pr,tag=@josler" # show me all of the items tagged "#pr" and "@josler"
```

The format of this filter is: `{type_of_filter}={value}`, with commas `,` separating each filter. The supported types of filter at this time are: `tag`, `status`, and `time`. The tag and status values are self-explanatory, and the value for a time filter is of the time format specified above.

Currently, these filters are an AND filter - they must all be true. Furthermore, the only supported operator is `=`.

Please note this is *mutually exclusive* with the flags for filtering by status directly, and the regular time filter as well.

### Auto

wdid has an `auto` command, where it can pull a list of potential items from various sources, and present them to the user. These can be selected to be turned into full items.

```
$ wdid auto
<type to filter>
<pick options with spacebar>
<hit enter to finish>
```

#### Auto-Github

This can be enabled by adding the relevant config:

```
[[auto]]
type = "github"
key = "accesstoken"
username = "my_username"
```

Options will then be sourced from Github, any issue or PR the author was involved in that was updated in $time. Issues and PR's that are closed will be auto-marked as "done".

### Auto-GoogleCalendar (WIP)

Enable with the following:

```
[[auto]]
type = "calendar"
username = "jamie@intercom.io"
```

Where the username is the name of your calendar you want to draw events from. Events that have taken place in the past will create items marked as "done".

Currently this requires a Google Calendar OAuth client secret file placed in `~/.config/wdid/client_secret.json`. You can obtain one of these from Google by setting up a project and enabling calendar API for it. In the future, wdid may provide this.

The first time this runs, it will ask the user for access through OAuth to the calendar, and then save the granted token locally.

### Configuration

Wdid should work out of the box with some sensible defaults. On first run it will populate a configuration file under `~/.config/wdid/config.toml`. This, by default, sets local storage up using [boltdb](https://github.com/coreos/bbolt).

#### Cross-Device Syncing

Currently, the suggested way to do this is to change the config file to point the store to somewhere that gets synced via an external method. For example, Dropbox works well:

```toml
[store]
type = "bolt"
file = "~/Dropbox/wdid.db"
```
