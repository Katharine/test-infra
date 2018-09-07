package cov

import (
	"errors"
	"fmt"
	"golang.org/x/tools/cover"
	"io"
	"sort"
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

func MergeProfiles(a []*cover.Profile, b []*cover.Profile) ([]*cover.Profile, error) {
	var result []*cover.Profile
	files := make(map[string]*cover.Profile, len(a))
	for _, profile := range a {
		// deep copy, so we don't modify the original
		np := copyProfile(*profile)
		result = append(result, &np)
		files[np.FileName] = &np
	}

	needsSort := false
	// Now merge b into the result
	for _, profile := range b {
		dest, ok := files[profile.FileName]
		if ok {
			// for a file that already exists, we assume it has the same blocks in the same order.
			if len(profile.Blocks) != len(dest.Blocks) {
				return nil, fmt.Errorf("numbers of blocks in %s mismatch", profile.FileName)
			}
			for i, block := range profile.Blocks {
				db := &dest.Blocks[i]
				if !blocksEqual(block, *db) {
					return nil, errors.New("coverage block mismatch")
				}
				db.Count += block.Count
			}
		} else {
			np := copyProfile(*profile)
			files[np.FileName] = &np
			result = append(result, &np)
			needsSort = true
		}
	}
	if needsSort {
		sort.Slice(result, func(i, j int) bool { return result[i].FileName < result[j].FileName })
	}
	return result, nil
}

func MergeMultipleProfiles(profiles [][]*cover.Profile) ([]*cover.Profile, error) {
	if len(profiles) < 1 {
		return nil, errors.New("can't merge zero profiles")
	}
	result := profiles[0]
	for _, profile := range profiles[1:] {
		var err error
		if result, err = MergeProfiles(result, profile); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func copyProfile(profile cover.Profile) cover.Profile {
	p := profile
	p.Blocks = make([]cover.ProfileBlock, len(profile.Blocks))
	copy(p.Blocks, profile.Blocks)
	return p
}

func blocksEqual(a cover.ProfileBlock, b cover.ProfileBlock) bool {
	return a.StartCol == b.StartCol && a.StartLine == b.StartLine &&
		a.EndCol == b.EndCol && a.EndLine == b.EndLine && a.NumStmt == b.NumStmt
}