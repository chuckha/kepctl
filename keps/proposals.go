package keps

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Proposals []*Proposal

type ByCreationDate Proposals

func (b ByCreationDate) Len() int           { return len(b) }
func (b ByCreationDate) Less(i, j int) bool { return b[i].CreationDate.After(b[j].CreationDate) }
func (b ByCreationDate) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

type ByTitle Proposals

func (b ByTitle) Len() int           { return len(b) }
func (b ByTitle) Less(i, j int) bool { return b[i].Title > b[j].Title }
func (b ByTitle) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

func (p *Proposals) AddProposal(proposal *Proposal) {
	*p = append(*p, proposal)
}
func (p *Proposals) SortBy(field string) {
	switch field {
	case "created", "creation", "creationDate":
		sort.Sort(ByCreationDate(*p))
	case "title":
		sort.Sort(ByTitle(*p))
	}
}

type Proposal struct {
	Title             string
	Authors           []string  `yaml:,flow`
	OwningSIG         string    `yaml:"owning-sig"`
	ParticipatingSIGs []string  `yaml:"participating-sigs",flow`
	Reviewers         []string  `yaml:,flow`
	Approvers         []string  `yaml:,flow`
	CreationDate      time.Time `yaml:"creation-date"`
	LastUpdated       time.Time `yaml:"last-updated"`
	Status            string
	SeeAlso           []string `yaml:"see-also"`

	Filename string `yaml:"-"`
}

func (p *Proposal) Filter(key, value string) bool {
	switch key {
	case "author":
		return Contains(p.Authors, value)
	case "status":
		return p.Status == value
	}
	return false
}

func Contains(haystack []string, needle string) bool {
	for _, hay := range haystack {
		if hay == needle {
			return true
		}
	}
	return false
}

type Parser struct {
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(in io.Reader) (*Proposal, error) {
	scanner := bufio.NewScanner(in)
	count := 0
	metadata := []byte{}
	for scanner.Scan() {
		line := scanner.Text() + "\n"
		if strings.Contains(line, "---") {
			count++
			continue
		}
		if count == 2 {
			break
		}
		metadata = append(metadata, []byte(line)...)

	}
	if err := scanner.Err(); err != nil {
		return nil, errors.WithStack(err)
	}
	proposal := &Proposal{}
	err := yaml.Unmarshal(metadata, proposal)
	return proposal, errors.WithStack(err)
}

var validProposalStatus = []string{"provisional", "implementable", "implemented", "deferred", "rejected", "withdrawn", "replaced"}

func Validate(p *Proposal) []error {
	var field fieldValidator
	field.isNonEmpty("title", p.Title)
	field.isNonEmptySlice("authors list", p.Authors)
	field.isNonEmpty("owning-sig", p.OwningSIG)
	field.isNonEmptySlice("reviewers list", p.Reviewers)
	field.isNonEmptySlice("approvers list", p.Approvers)
	field.isNonZeroTime("creation date", p.CreationDate)
	field.isNonZeroTime("last updated date", p.LastUpdated)
	if !p.LastUpdated.IsZero() && !p.CreationDate.IsZero() {
		field.isAfter("last updated date", "creation date", p.LastUpdated, p.CreationDate)
	}
	field.isNonEmpty("status", p.Status)
	if p.Status != "" {
		field.isOneOf("status", p.Status, validProposalStatus)
	}
	return []error(field)
}

type fieldValidator []error

func (fv *fieldValidator) isNonEmpty(field, value string) {
	if value == "" {
		*fv = append(*fv, fmt.Errorf("%s cannot be empty", field))
	}
}

func (fv *fieldValidator) isNonEmptySlice(field string, value []string) {
	if len(value) == 0 {
		*fv = append(*fv, fmt.Errorf("%s cannot be empty", field))
	}
}

func (fv *fieldValidator) isNonZeroTime(field string, t time.Time) {
	if t.IsZero() {
		*fv = append(*fv, fmt.Errorf("%s cannot be empty", field))
	}
}

func (fv *fieldValidator) isOneOf(field, value string, validValues []string) {
	for _, v := range validValues {
		if value == v {
			return
		}
	}
	*fv = append(*fv, fmt.Errorf("'%s' is not a valid %s. Valid options are '%s'", value, field, strings.Join(validValues, "', '")))
}

func (fv *fieldValidator) isAfter(field1, field2 string, value1, value2 time.Time) {
	if !value1.After(value2) {
		*fv = append(*fv, fmt.Errorf("%s must be later than %s", field1, field2))
	}
}
