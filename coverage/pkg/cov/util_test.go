package cov_test

import (
	"bytes"
	. "github.com/onsi/gomega"
	"golang.org/x/tools/cover"
	"k8s.io/test-infra/coverage/pkg/cov"
	"testing"
)

func TestDumpProfileOneFile(t *testing.T) {
	RegisterTestingT(t)
	a := []*cover.Profile{
		{
			FileName: "foo.go",
			Mode:     "count",
			Blocks: []cover.ProfileBlock{
				{StartLine: 1, StartCol: 3, EndLine: 20, EndCol: 1, NumStmt: 10, Count: 3},
				{StartLine: 22, StartCol: 5, EndLine: 28, EndCol: 2, NumStmt: 5, Count: 2},
			},
		},
	}
	var buffer bytes.Buffer
	cov.DumpProfile(a, &buffer)
	Expect(buffer.String()).To(Equal(`mode: count
foo.go:1.3,20.1 10 3
foo.go:22.5,28.2 5 2
`))
}

func TestDumpProfileMultipleFiles(t *testing.T) {
	RegisterTestingT(t)
	a := []*cover.Profile{
		{
			FileName: "bar.go",
			Mode:     "atomic",
			Blocks: []cover.ProfileBlock{
				{StartLine: 5, StartCol: 1, EndLine: 16, EndCol: 7, NumStmt: 7, Count: 0},
			},
		},
		{
			FileName: "foo.go",
			Mode:     "atomic",
			Blocks: []cover.ProfileBlock{
				{StartLine: 1, StartCol: 3, EndLine: 20, EndCol: 1, NumStmt: 10, Count: 3},
				{StartLine: 22, StartCol: 5, EndLine: 28, EndCol: 2, NumStmt: 5, Count: 2},
			},
		},
	}
	var buffer bytes.Buffer
	cov.DumpProfile(a, &buffer)
	Expect(buffer.String()).To(Equal(`mode: atomic
bar.go:5.1,16.7 7 0
foo.go:1.3,20.1 10 3
foo.go:22.5,28.2 5 2
`))
}
