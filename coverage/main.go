package main

import (
	"log"
	"github.com/spf13/cobra"
	"k8s.io/test-infra/coverage/cmd/diff"
	"k8s.io/test-infra/coverage/cmd/merge"
)

var rootCommand = &cobra.Command{
	Use: "coverage",
	Short: "coverage is a tool for manipulating Go coverage files.",
}

func run() error {
	rootCommand.AddCommand(diff.MakeCommand())
	rootCommand.AddCommand(merge.MakeCommand())
	return rootCommand.Execute()
}


func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
