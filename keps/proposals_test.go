package keps_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/chuckha/kepview/keps"
)

func TestValidParsing(t *testing.T) {
	testcases := []struct {
		name         string
		fileContents string
	}{
		{
			"simple test",
			`---
title: test
authors:
  - "@jpbetz"
  - "@roycaihw"
  - "@sttts"
owning-sig: sig-api-machinery
participating-sigs:
  - sig-api-machinery
  - sig-architecture
reviewers:
  - "@deads2k"
  - "@lavalamp"
  - "@liggitt"
  - "@mbohlool"
  - "@sttts"
approvers:
  - "@deads2k"
  - "@lavalamp"
creation-date: 2018-04-15
last-updated: 2018-04-24
status: provisional
---`,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := keps.NewParser()
			contents := strings.NewReader(tc.fileContents)
			out, err := p.Parse(contents)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			if out == nil {
				t.Fatal("out should not be nil")
			}
		})
	}
}

func TestProposalValidation(t *testing.T) {
	testcases := []struct {
		name           string
		content        string
		expectedErrors []error
	}{
		{
			name: "missing title",
			content: `---
#title: test
authors:
  - "@jpbetz"
owning-sig: sig-api-machinery
reviewers:
  - "@deads2k"
approvers:
  - "@deads2k"
creation-date: 2018-04-15
last-updated: 2018-04-24
status: provisional
`,
			expectedErrors: []error{fmt.Errorf("title cannot be empty")},
		},
		{
			name: "missing authors",
			content: `---
title: test
#authors:
#  - "@jpbetz"
owning-sig: sig-api-machinery
reviewers:
  - "@deads2k"
approvers:
  - "@deads2k"
creation-date: 2018-04-15
last-updated: 2018-04-24
status: provisional
`,
			expectedErrors: []error{fmt.Errorf("authors list cannot be empty")},
		},
		{
			name: "missing owning-sig",
			content: `---
title: test
authors:
- "@jpbetz"
#owning-sig: sig-api-machinery
reviewers:
  - "@deads2k"
approvers:
  - "@deads2k"
creation-date: 2018-04-15
last-updated: 2018-04-24
status: provisional
`,
			expectedErrors: []error{fmt.Errorf("owning-sig cannot be empty")},
		},
		{
			name: "missing reviewers",
			content: `---
title: test
authors:
- "@jpbetz"
owning-sig: sig-api-machinery
#reviewers:
#  - "@deads2k"
approvers:
  - "@deads2k"
creation-date: 2018-04-15
last-updated: 2018-04-24
status: provisional
`,
			expectedErrors: []error{fmt.Errorf("reviewers list cannot be empty")},
		},
		{
			name: "missing approvers",
			content: `---
title: test
authors:
- "@jpbetz"
owning-sig: sig-api-machinery
reviewers:
  - "@deads2k"
#approvers:
#  - "@deads2k"
creation-date: 2018-04-15
last-updated: 2018-04-24
status: provisional
`,
			expectedErrors: []error{fmt.Errorf("approvers list cannot be empty")},
		},
		{
			name: "missing creation date",
			content: `---
title: test
authors:
- "@jpbetz"
owning-sig: sig-api-machinery
reviewers:
  - "@deads2k"
approvers:
  - "@deads2k"
#creation-date: 2018-04-15
last-updated: 2018-04-24
status: provisional
`,
			expectedErrors: []error{fmt.Errorf("creation date cannot be empty")},
		},
		{
			name: "missing last updated",
			content: `---
title: test
authors:
- "@jpbetz"
owning-sig: sig-api-machinery
reviewers:
  - "@deads2k"
approvers:
  - "@deads2k"
creation-date: 2018-04-15
#last-updated: 2018-04-24
status: provisional
`,
			expectedErrors: []error{fmt.Errorf("last updated date cannot be empty")},
		},
		{
			name: "last updated date is before creation date",
			content: `---
title: test
authors:
- "@jpbetz"
owning-sig: sig-api-machinery
reviewers:
  - "@deads2k"
approvers:
  - "@deads2k"
creation-date: 2018-04-15
last-updated: 2018-04-14
status: provisional
`,
			expectedErrors: []error{fmt.Errorf("last updated date must be later than creation date")},
		},
		{
			name: "missing status",
			content: `---
title: test
authors:
- "@jpbetz"
owning-sig: sig-api-machinery
reviewers:
  - "@deads2k"
approvers:
  - "@deads2k"
creation-date: 2018-04-15
last-updated: 2018-04-24
#status: provisional
`,
			expectedErrors: []error{fmt.Errorf("status cannot be empty")},
		},
		{
			name: "invalid status",
			content: `---
title: test
authors:
- "@jpbetz"
owning-sig: sig-api-machinery
reviewers:
  - "@deads2k"
approvers:
  - "@deads2k"
creation-date: 2018-04-15
last-updated: 2018-04-24
status: foo
`,
			expectedErrors: []error{fmt.Errorf("'foo' is not a valid status. Valid options are 'provisional', 'implementable', 'implemented', 'deferred', 'rejected', 'withdrawn', 'replaced'")},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := keps.NewParser()
			kep, err := p.Parse(strings.NewReader(tc.content))
			if err != nil {
				t.Errorf("error parsing proposal: %v", err)
			}
			got := keps.Validate(kep)
			if !reflect.DeepEqual(got, tc.expectedErrors) {
				t.Errorf("expected errors:\n%v\ngot:\n%v", tc.expectedErrors, got)
			}
		})

	}
}
