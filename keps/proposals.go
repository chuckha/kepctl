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

func Validate(p *Proposal) []error {
	var e []error
	if p.Title == "" {
		e = append(e, fmt.Errorf("title cannot be empty"))
	}
	if len(p.Authors) == 0 {
		e = append(e, fmt.Errorf("authors list cannot be empty"))
	}
	if p.OwningSIG == "" {
		e = append(e, fmt.Errorf("owning-sig cannot be empty"))
	}
	if len(p.Reviewers) == 0 {
		e = append(e, fmt.Errorf("reviewers list cannot be empty"))
	}
	if len(p.Approvers) == 0 {
		e = append(e, fmt.Errorf("approvers list cannot be empty"))
	}
	if p.CreationDate.IsZero() {
		e = append(e, fmt.Errorf("creation date cannot be empty"))
	}
	if p.LastUpdated.IsZero() {
		e = append(e, fmt.Errorf("last updated date cannot be empty"))
	}
	if !p.CreationDate.IsZero() && !p.LastUpdated.IsZero() && p.CreationDate.After(p.LastUpdated) {
		e = append(e, fmt.Errorf("last updated date must be later than creation date"))
	}
	if p.Status == "" {
		e = append(e, fmt.Errorf("status cannot be empty"))
	}
	if p.Status != "" && !isValidStatus(p.Status) {
		e = append(e, fmt.Errorf("'%s' is not a valid status. The valid statuses are '%s'", p.Status, strings.Join(validProposalStatus, "', '")))
	}
	return e
}

var validProposalStatus = []string{"provisional", "implementable", "implemented", "deferred", "rejected", "withdrawn", "replaced"}

func isValidStatus(status string) bool {
	for _, s := range validProposalStatus {
		if s == status {
			return true
		}
	}
	return false
}
