package cov

import (
	"errors"
	"fmt"
	"golang.org/x/tools/cover"
	"io"
)

// DumpProfile dumps the profiles given to writer in go coverage format.
func DumpProfile(profiles []*cover.Profile, writer io.Writer) error {
	if len(profiles) == 0 {
		return errors.New("can't write an empty profiles")
	}
	if _, err := io.WriteString(writer, "mode: "+profiles[0].Mode+"\n"); err != nil {
		return err
	}
	for _, profile := range profiles {
		for _, block := range profile.Blocks {
			if _, err := fmt.Fprintf(writer, "%s:%d.%d,%d.%d %d %d\n", profile.FileName, block.StartLine, block.StartCol, block.EndLine, block.EndCol, block.NumStmt, block.Count); err != nil {
				return err
			}
		}
	}
	return nil
}

func DiffProfiles(before []*cover.Profile, after []*cover.Profile) ([]*cover.Profile, error) {
	diff := []*cover.Profile{}
	for i, beforeProfile := range before {
		afterProfile := after[i]
		if beforeProfile.FileName != afterProfile.FileName {
			return nil, errors.New("coverage filename mismatch")
		}
		diffProfile := cover.Profile{FileName: beforeProfile.FileName, Mode: beforeProfile.Mode}
		for j, beforeBlock := range beforeProfile.Blocks {
			afterBlock := afterProfile.Blocks[j]
			if beforeBlock.StartLine != afterBlock.StartLine || beforeBlock.EndLine != afterBlock.EndLine {
				return nil, errors.New("coverage block mismatch")
			}
			diffBlock := cover.ProfileBlock{
				StartLine: beforeBlock.StartLine,
				StartCol:  beforeBlock.StartCol,
				EndLine:   beforeBlock.EndLine,
				EndCol:    beforeBlock.EndCol,
				NumStmt:   beforeBlock.NumStmt,
				Count:     afterBlock.Count - beforeBlock.Count,
			}
			diffProfile.Blocks = append(diffProfile.Blocks, diffBlock)
		}
		diff = append(diff, &diffProfile)
	}
	return diff, nil
}