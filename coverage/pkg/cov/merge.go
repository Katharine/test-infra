package cov

import (
	"golang.org/x/tools/cover"
	"sort"
	"errors"
	"fmt"
)

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
