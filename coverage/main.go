package main

import (
	"golang.org/x/tools/cover"
	"./pkg/cov"
	"os"
)

func main() {
	before, err := cover.ParseProfiles(os.Args[1])
	if err != nil {
		panic(err)
	}
	after, err := cover.ParseProfiles(os.Args[2])
	if err != nil {
		panic(err)
	}
	diff, err := cov.DiffProfiles(before, after)
	if err := cov.DumpProfile(diff, os.Stdout); err != nil {
		panic(err)
	}
}
