package main

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/chuckha/kepview/keps"
)

type info struct {
	name string
}

func (i *info) Name() string       { return i.name }
func (i *info) Size() int64        { return 0 }
func (i *info) Mode() os.FileMode  { return os.FileMode(100) }
func (i *info) ModTime() time.Time { return time.Date(2019, 4, 20, 0, 0, 0, 0, nil) }
func (i *info) IsDir() bool        { return false }
func (i *info) Sys() interface{}   { return struct{}{} }

type myparser struct {
	proposal *keps.Proposal
}

func (p *myparser) Parse(reader io.Reader) (*keps.Proposal, error) {
	return p.proposal, nil
}

type myopener struct {
	file *os.File
}

func (o *myopener) Open(path string) (*os.File, error) {
	return o.file, nil
}

type mylogger struct{}

func (l *mylogger) Debugf(format string, args ...interface{}) {}

func defaultTestEnhancementFinder(...finderOpts) *EnhancementFinder {
	return &EnhancementFinder{
		opener:          &myopener{},
		parser:          &myparser{},
		log:             &mylogger{},
		filenameFilters: defaultFilters(),
	}
}

func TestFindEnhancementsIgnores(t *testing.T) {
	testcases := []struct {
		name     string
		filename string
	}{
		{
			"a basic readme",
			"README.md",
		},
		{
			"owners file",
			"OWNERS",
		},
		{
			"images",
			"something.png",
		},
		{
			"ignore templates",
			"myfavorite-template.md",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ef := defaultTestEnhancementFinder()
			out := &keps.Proposals{}
			fe := ef.Find(out)
			i := &info{tc.filename}
			if err := fe("test", i, nil); err != nil {
				t.Fatalf("%+v", err)
			}
			if len(*out) != 0 {
				t.Fatalf("Did not expect to find anything but found %v", out)
			}
		})
	}
}

func TestEnhancementFinder(t *testing.T) {
	testcases := []struct {
		name     string
		filename string
	}{
		{
			"simple test",
			"my-simple-test.md",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ef := defaultTestEnhancementFinder()
			ef.parser = &myparser{&keps.Proposal{}}
			out := &keps.Proposals{}
			fe := ef.Find(out)
			i := &info{tc.filename}
			if err := fe("test", i, nil); err != nil {
				t.Fatalf("%+v", err)
			}
			if len(*out) != 1 {
				t.Fatalf("Expected 1 item but found: %v", out)
			}
			if (*out)[0].Filename != "test" {
				t.Fatalf("expected proposal to have a filename of %q but had %q", tc.filename, (*out)[0].Filename)
			}
		})
	}
}
