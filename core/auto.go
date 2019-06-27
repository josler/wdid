package core

import (
	"context"
	"sort"
	"time"

	"github.com/AlecAivazis/survey"
	"github.com/josler/wdid/auto"
)

type AutoSource interface {
	Precheck()
	Load(startTime, endTime time.Time) []*auto.Option
}

type AutoConf interface {
	AutoType() string
	AuthKey() string
	AutoUsername() string
}

type autoLoader struct {
	ctx    context.Context
	picker picker
}

func Auto(ctx context.Context, timeString string, confs ...AutoConf) error {
	store := ctx.Value("store").(Store)
	at, err := TimeParser{Input: timeString}.Parse()
	if err != nil {
		return err
	}

	loader := &autoLoader{ctx: ctx, picker: &autoPicker{}}

	pickedOptions := loader.loadOptions(at, confs...)
	savedItems := []*Item{}

	ic := ItemCreator{ctx: ctx}
	for _, o := range pickedOptions {
		item, err := ic.Create(o.Data(), o.DateTime().Local())
		if err != nil {
			continue
		}
		if o.Status() == "done" {
			item.Do()
			store.Save(item)
		}
		savedItems = append(savedItems, item)
	}

	NewItemPrinter(ctx).Print(savedItems...)

	return nil
}

func (loader *autoLoader) loadOptions(timespan *Timespan, confs ...AutoConf) []*auto.Option {
	options := []*auto.Option{}
	ch := make(chan []*auto.Option)

	for _, c := range confs {
		loader.sourceFor(c).Precheck()
	}

	for _, c := range confs {
		go loader.loadToChannel(ch, timespan, c)
	}

	i := 0
OuterLoop:
	for i < len(confs) {
		select {
		case loaded := <-ch:
			options = append(options, loaded...)
			i += 1
		case <-time.After(5 * time.Second):
			break OuterLoop
		}
	}

	sort.Slice(options, func(i, j int) bool {
		return options[i].DateTime().Before(options[j].DateTime())
	})

	return loader.picker.Pick(options)
}

func (loader *autoLoader) sourceFor(conf AutoConf) AutoSource {
	switch conf.AutoType() {
	case "github":
		return auto.NewGithubClient(loader.ctx, conf.AutoUsername(), conf.AuthKey())
	case "calendar":
		return auto.NewGoogleCalendar(conf.AutoUsername())
	}
	return nil
}

func (loader *autoLoader) loadToChannel(ch chan<- []*auto.Option, timespan *Timespan, c AutoConf) {
	ch <- loader.loadAutoSource(timespan, loader.sourceFor(c))
}

func (loader *autoLoader) loadAutoSource(timespan *Timespan, source AutoSource) []*auto.Option {
	if source == nil {
		return []*auto.Option{}
	}
	return source.Load(timespan.Start, timespan.End)
}

type picker interface {
	Pick(options []*auto.Option) []*auto.Option
}

type autoPicker struct{}

func (picker *autoPicker) Pick(options []*auto.Option) []*auto.Option {
	optionStrings := []string{}
	for _, opt := range options {
		optionStrings = append(optionStrings, picker.trimData(opt))
	}

	chosen := []string{}
	prompt := &survey.MultiSelect{
		Message:  "Pick:",
		Options:  optionStrings,
		PageSize: 20,
	}
	survey.AskOne(prompt, &chosen, nil)

	pickedOptions := []*auto.Option{}
	// we won't have many options so this is OK (low n)
	for _, c := range chosen {
		for _, o := range options {
			if c == picker.trimData(o) {
				pickedOptions = append(pickedOptions, o)
			}
		}
	}

	return pickedOptions
}

func (picker *autoPicker) trimData(opt *auto.Option) string {
	return TrimString(opt.Data(), 10)
}
