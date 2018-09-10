package cov_test

import (
	. "github.com/onsi/gomega"
	"golang.org/x/tools/cover"
	"k8s.io/test-infra/coverage/pkg/cov"
	"testing"
)

func TestMergeProfilesSimilar(t *testing.T) {
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
	Expect(cov.MergeProfiles(a, b)).To(Equal([]*cover.Profile{
		{
			FileName: "a.go",
			Mode:     "count",
			Blocks: []cover.ProfileBlock{
				{StartLine: 1, StartCol: 14, EndLine: 5, EndCol: 13, NumStmt: 4, Count: 10},
				{StartLine: 7, StartCol: 4, EndLine: 12, EndCol: 4, NumStmt: 3, Count: 4},
			},
		},
	}))
}

func TestMergeProfilesDisjoint(t *testing.T) {
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
			FileName: "b.go",
			Mode:     "count",
			Blocks: []cover.ProfileBlock{
				{StartLine: 1, StartCol: 14, EndLine: 5, EndCol: 13, NumStmt: 4, Count: 7},
				{StartLine: 7, StartCol: 4, EndLine: 12, EndCol: 4, NumStmt: 3, Count: 2},
			},
		},
	}
	Expect(cov.MergeProfiles(a, b)).To(Equal([]*cover.Profile{
		{
			FileName: "a.go",
			Mode:     "count",
			Blocks: []cover.ProfileBlock{
				{StartLine: 1, StartCol: 14, EndLine: 5, EndCol: 13, NumStmt: 4, Count: 3},
				{StartLine: 7, StartCol: 4, EndLine: 12, EndCol: 4, NumStmt: 3, Count: 2},
			},
		},
		{
			FileName: "b.go",
			Mode:     "count",
			Blocks: []cover.ProfileBlock{
				{StartLine: 1, StartCol: 14, EndLine: 5, EndCol: 13, NumStmt: 4, Count: 7},
				{StartLine: 7, StartCol: 4, EndLine: 12, EndCol: 4, NumStmt: 3, Count: 2},
			},
		},
	}))
}

func TestMergeProfilesOverlapping(t *testing.T) {
	RegisterTestingT(t)
	a := []*cover.Profile{
		{
			FileName: "bar.go",
			Mode:     "count",
			Blocks: []cover.ProfileBlock{
				{StartLine: 1, StartCol: 3, EndLine: 20, EndCol: 1, NumStmt: 10, Count: 17},
			},
		},
		{
			FileName: "foo.go",
			Mode:     "count",
			Blocks: []cover.ProfileBlock{
				{StartLine: 1, StartCol: 14, EndLine: 5, EndCol: 13, NumStmt: 4, Count: 3},
				{StartLine: 7, StartCol: 4, EndLine: 12, EndCol: 4, NumStmt: 3, Count: 2},
			},
		},
	}
	b := []*cover.Profile{
		{
			FileName: "bar.go",
			Mode:     "count",
			Blocks: []cover.ProfileBlock{
				{StartLine: 1, StartCol: 3, EndLine: 20, EndCol: 1, NumStmt: 10, Count: 8},
			},
		},
		{
			FileName: "baz.go",
			Mode:     "count",
			Blocks: []cover.ProfileBlock{
				{StartLine: 1, StartCol: 14, EndLine: 5, EndCol: 13, NumStmt: 4, Count: 7},
				{StartLine: 7, StartCol: 4, EndLine: 12, EndCol: 4, NumStmt: 3, Count: 2},
			},
		},
	}
	Expect(cov.MergeProfiles(a, b)).To(Equal([]*cover.Profile{
		{
			FileName: "bar.go",
			Mode:     "count",
			Blocks: []cover.ProfileBlock{
				{StartLine: 1, StartCol: 3, EndLine: 20, EndCol: 1, NumStmt: 10, Count: 25},
			},
		},
		{
			FileName: "baz.go",
			Mode:     "count",
			Blocks: []cover.ProfileBlock{
				{StartLine: 1, StartCol: 14, EndLine: 5, EndCol: 13, NumStmt: 4, Count: 7},
				{StartLine: 7, StartCol: 4, EndLine: 12, EndCol: 4, NumStmt: 3, Count: 2},
			},
		},
		{
			FileName: "foo.go",
			Mode:     "count",
			Blocks: []cover.ProfileBlock{
				{StartLine: 1, StartCol: 14, EndLine: 5, EndCol: 13, NumStmt: 4, Count: 3},
				{StartLine: 7, StartCol: 4, EndLine: 12, EndCol: 4, NumStmt: 3, Count: 2},
			},
		},
	}))
}

func TestMergeMultipleProfiles(t *testing.T) {
	a := []*cover.Profile{
		{
			FileName: "foo.go",
			Mode:     "count",
			Blocks: []cover.ProfileBlock{
				{StartLine: 1, StartCol: 3, EndLine: 20, EndCol: 1, NumStmt: 10, Count: 3},
			},
		},
	}
	b := []*cover.Profile{
		{
			FileName: "foo.go",
			Mode:     "count",
			Blocks: []cover.ProfileBlock{
				{StartLine: 1, StartCol: 3, EndLine: 20, EndCol: 1, NumStmt: 10, Count: 5},
			},
		},
	}
	c := []*cover.Profile{
		{
			FileName: "foo.go",
			Mode:     "count",
			Blocks: []cover.ProfileBlock{
				{StartLine: 1, StartCol: 3, EndLine: 20, EndCol: 1, NumStmt: 10, Count: 8},
			},
		},
	}
	Expect(cov.MergeMultipleProfiles([][]*cover.Profile{a, b, c})).To(Equal([]*cover.Profile{
		{
			FileName: "foo.go",
			Mode:     "count",
			Blocks: []cover.ProfileBlock{
				{StartLine: 1, StartCol: 3, EndLine: 20, EndCol: 1, NumStmt: 10, Count: 16},
			},
		},
	}))
}
