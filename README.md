### Overview

What Did I Do (wdid) is a small CLI tool to track what you have been working on. You can `add`, `list`, `edit`, `do`, `skip`, `bump`, `show` and `rm` items. Alongside this manual tracking, there's an `auto` feature that can help automate fetching information about what you've done.

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
```

### Installation

```
go get -u gitlab.com/josler/wdid/...
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

You can filter by type too:

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
- `YYYY-MM-DD` (`2006-01-02` in Go time format): Start of given day in your TZ. Ends end of that day.
- `YYYY-MM-DDTHH:MM`: particular time on a day in your TZ. Ends end of that day.

Note that these times cover a _range_ of values. Usually from the start of the indicated day (00:00) to the end of the day (23:59) at the end of the period, inclusive.

When adding items, or setting the time for an item, wdid uses the _start_ of the period to do so. When searching for items, wdid uses the range. This sounds more complicated than it is, in practise it does what you'd expect.

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

### Configuration

Wdid should work out of the box with some sensible defaults. On first run it will populate a configuration file under `~/.config/wdid/config.toml`. This, by default, sets local storage up using [boltdb](https://github.com/coreos/bbolt).

#### Cross-Device Syncing

Currently, the suggested way to do this is to change the config file to point the store to somewhere that gets synced via an external method. For example, Dropbox works well:

```toml
[store]
type = "bolt"
file = "~/Dropbox/wdid.db"
```
