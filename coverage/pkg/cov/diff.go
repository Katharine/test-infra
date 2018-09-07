package cov

import (
	"golang.org/x/tools/cover"
	"errors"
)

func DiffProfiles(before []*cover.Profile, after []*cover.Profile) ([]*cover.Profile, error) {
	var diff []*cover.Profile
	for i, beforeProfile := range before {
		afterProfile := after[i]
		if beforeProfile.FileName != afterProfile.FileName {
			return nil, errors.New("coverage filename mismatch")
		}
		diffProfile := cover.Profile{FileName: beforeProfile.FileName, Mode: beforeProfile.Mode}
		for j, beforeBlock := range beforeProfile.Blocks {
			afterBlock := afterProfile.Blocks[j]
			if !blocksEqual(beforeBlock, afterBlock) {
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
