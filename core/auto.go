package core

import (
	"context"
	"time"

	"github.com/AlecAivazis/survey"
	"gitlab.com/josler/wdid/auto"
)

type AutoSource interface {
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

	for _, o := range pickedOptions {
		item := NewItem(o.Data(), o.DateTime().Local())
		if o.Status() == "done" {
			item.Do()
		}
		err := store.Save(item)
		if err == nil {
			savedItems = append(savedItems, item)
		}
	}

	NewItemPrinter(ctx).Print(savedItems...)

	return nil
}

func (loader *autoLoader) loadOptions(timespan *Timespan, confs ...AutoConf) []*auto.Option {
	options := []*auto.Option{}
	for _, c := range confs {
		options = append(options, loader.loadAutoSource(timespan, loader.sourceFor(c))...)
	}

	return loader.picker.Pick(options)
}

func (loader *autoLoader) sourceFor(conf AutoConf) AutoSource {
	switch conf.AutoType() {
	case "github":
		return auto.NewGithubClient(loader.ctx, conf.AutoUsername(), conf.AuthKey())
	}
	return nil
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
		optionStrings = append(optionStrings, opt.Data())
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
			if c == o.Data() {
				pickedOptions = append(pickedOptions, o)
			}
		}
	}

	return pickedOptions
}
