package parser

import (
	"fmt"
	"strings"

	"github.com/josler/wdid/filter"
)

type ToFilterFn func(comparison filter.FilterComparison, val string) (filter.Filter, error)

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
	trimmedIdentifier := strings.Trim(identifier.val, " ")
	filterFn, ok := p.filterFnMap[trimmedIdentifier]
	if !ok {
		return fmt.Errorf("failed to parse, unrecognized filter: %q", identifier.val)
	}
	comparison, ok := <-p.itemchan // drain the comparison
	if !ok {
		return fmt.Errorf("failed to parse %q, missing comparison", identifier.val)
	}
	valueItem, ok := <-p.itemchan // next is the valueItem
	if !ok || valueItem.typ != lexItemString {
		return fmt.Errorf("failed to parse %q, missing value", identifier.val)
	}

	if p.results == nil {
		p.results = []filter.Filter{}
	}

	var filterComparison filter.FilterComparison
	switch comparison.typ {
	case lexItemEq:
		filterComparison = filter.FilterEq
	case lexItemNe:
		filterComparison = filter.FilterNe
	case lexItemGt:
		filterComparison = filter.FilterGt
	case lexItemLt:
		filterComparison = filter.FilterLt
	}

	trimmedValue := strings.Trim(valueItem.val, " ")
	result, err := filterFn(filterComparison, trimmedValue)
	if err != nil {
		return err
	}
	p.results = append(p.results, result)
	return nil
}
