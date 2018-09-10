package cov_test

import (
	. "github.com/onsi/gomega"
	"golang.org/x/tools/cover"
	"k8s.io/test-infra/coverage/pkg/cov"
	"testing"
)

func TestDiffProfilesBasicDiff(t *testing.T) {
	RegisterTestingT(t)
	a := []*cover.Profile{
		{
			FileName: "a.go",
			Mode:     "count",
			Blocks: []cover.ProfileBlock{
				{StartLine: 1, StartCol: 14, EndLine: 5, EndCol: 13, NumStmt: 4, Count: 3},
				{StartLine: 7, StartCol: 4, EndLine: 12, EndCol: 4, NumStmt: 3, Count: 2},
			},
		},
	}
	b := []*cover.Profile{
		{
			FileName: "a.go",
			Mode:     "count",
			Blocks: []cover.ProfileBlock{
				{StartLine: 1, StartCol: 14, EndLine: 5, EndCol: 13, NumStmt: 4, Count: 7},
				{StartLine: 7, StartCol: 4, EndLine: 12, EndCol: 4, NumStmt: 3, Count: 2},
			},
		},
	}

	Expect(cov.DiffProfiles(a, b)).To(Equal([]*cover.Profile{
		{
			FileName: "a.go",
			Mode:     "count",
			Blocks: []cover.ProfileBlock{
				{StartLine: 1, StartCol: 14, EndLine: 5, EndCol: 13, NumStmt: 4, Count: 4},
				{StartLine: 7, StartCol: 4, EndLine: 12, EndCol: 4, NumStmt: 3, Count: 0},
			},
		},
	}))
}
