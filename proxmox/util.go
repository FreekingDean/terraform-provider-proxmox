package proxmox

import (
	"fmt"
	"strconv"
)

const (
	B int64 = 1 << (10 * iota)
	K
	M
	G
	T
)

var (
	sizeMap = map[string]int64{
		"B": B,
		"K": K,
		"M": M,
		"G": G,
		"T": T,
	}
)

func strToGB(in string) (int64, error) {
	denom := in[len(in)-1:]
	sizeStr := in[0 : len(in)-1]
	size32, err := strconv.Atoi(sizeStr)
	if err != nil {
		return 0, err
	}
	size := int64(size32)
	if _, ok := sizeMap[denom]; !ok {
		return 0, fmt.Errorf("Could not find denom %s", denom)
	}
	size = size * sizeMap[denom]
	return size / G, nil
}
