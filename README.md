[Documentation](https://j-osler.gitbook.io/wdid/)

### Overview

What Did I Do (wdid) is a small CLI tool to track what you have been working on. You can `add`, `list`, `edit`, `do`, `skip`, `bump`, `show` and `rm` items. There's tagging built in, and the ability to list and filter your items by tag, status, and time. Using these filters, you can organise your data however you like with groups. It's very fast and flexible. Here's a sample of what you can do (run it yourself to see the lovely colors!), though you can do an awful lot more; such as organising projects, areas of work, time frames, taking notes etc...

<img width="498" alt="terminal screen shot" src="https://user-images.githubusercontent.com/167061/68556700-4e199800-0401-11ea-9e81-cdf948663f62.png">

#### Add an item

```
$ wdid add "my item for #project"
⇒ w9hjba -- Fri, 09 Nov 2019 15:03:08
Tags: [#project]
Data:
my item for #project
```

#### List items with filters

```
$ wdid "tag=#project,status=waiting"

- Thu Nov 07
⇒ o2wjb9     send email to @josler about #project [@josler #project]

- Fri Nov 08
⇒ w9hjba     my item for #project                 [#project]
```

#### Take action on items

```
$ wdid do o2w
✔ o2wjb9 -- Thu, 07 Nov 2019 00:00:00
Tags: [@josler #project]
Data:
send email to @josler about #project
```

### Why command line?

I spend much of my time with a terminal open. It's right there, always a cmd-tab away. Being able to quickly record notes without having a website open, or having some app consuming memory is really useful. I also want to make sure that the data is easily accessible, and works with other tools wherever possible. Building on the command line enables that "for free". Your data is exportable, and we have different output formats for both humans to consume and for interop with various tools (json and structured text outputs!)

### Why personal?

This is a tool to track your personal to-do's, work done, notes, etc. It deliberately eschews complexity added by networking, sharing, and large-scale project management. In doing this it can remain, small, simple, fast, and useful.

### Installation

```
go get -u github.com/josler/wdid/...
```

Or check the [releases](https://github.com/josler/wdid/releases) page for prebuilt binaries.

### Usage

See [documentation](https://j-osler.gitbook.io/wdid/) for more.

```
$ wdid help
usage: wdid [<flags>] <command> [<args> ...]

A tool to track what you did.

Flags:
  -h, --help          Show context-sensitive help (also try --help-long and --help-man).
  -v, --verbose       Enable verbose logging.
      --format=human  format to print in ('human', 'text', or 'json).
      --version       Show application version.

Commands:
  help [<command>...]
  bump [<flags>] <id>
  add [<flags>] [<new-item>]
  do <id>
  edit [<flags>] <id> [<description>]
  group --name=NAME --filters=FILTERS
  group-rm --name=NAME
  group-ls
  import [<in>]
  ls* [<flags>] [<filters>]
  rm <id>
  skip <id>
  show <id>
  tag
  tag-ls
```
