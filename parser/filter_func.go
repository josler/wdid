package parser

import (
	"fmt"

	"github.com/josler/wdid/filter"
)

type ToFilterFn func(val string) (filter.Filter, error)

type Parser struct {
	filterFnMap map[string]ToFilterFn
	results     []filter.Filter
	itemchan    chan lexedItem
	err         error
}

func (p *Parser) RegisterToFilter(identifierName string, filterFn ToFilterFn) {
	if p.filterFnMap == nil {
		p.filterFnMap = map[string]ToFilterFn{}
	}
	p.filterFnMap[identifierName] = filterFn
}

func (p *Parser) Parse(input string) ([]filter.Filter, error) {
	_, p.itemchan = lex(input)
	p.parse()
	return p.results, p.err
}

func (p *Parser) parse() {
	for {
		i, ok := <-p.itemchan
		if !ok {
			// channel closed
			return
		}
		var err error

		switch i.typ {
		case lexItemIdentifier:
			err = p.parseIdentifier(i)
		}

		if err != nil {
			p.err = err
			return
		}
	}
}

func (p *Parser) parseIdentifier(identifier lexedItem) error {
	filterFn, ok := p.filterFnMap[identifier.val]
	if !ok {
		return fmt.Errorf("failed to parse, unrecognized filter: %q", identifier.val)
	}
	_, ok = <-p.itemchan // drain the comparison
	if !ok {
		return fmt.Errorf("failed to parse %q, missing eq", identifier.val)
	}
	valueItem, ok := <-p.itemchan // next is the valueItem
	if !ok || valueItem.typ != lexItemString {
		return fmt.Errorf("failed to parse %q, missing value", identifier.val)
	}

	if p.results == nil {
		p.results = []filter.Filter{}
	}
	result, err := filterFn(valueItem.val)
	if err != nil {
		return err
	}
	p.results = append(p.results, result)
	return nil
}
