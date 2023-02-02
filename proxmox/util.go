package proxmox

import (
	"crypto/rand"
	"fmt"
	"net"
	"strconv"
)

const (
	B int = 1 << (10 * iota)
	K
	M
	G
	T
)

var (
	sizeMap = map[string]int{
		"B": B,
		"K": K,
		"M": M,
		"G": G,
		"T": T,
	}
)

func strToGB(in string) (int, error) {
	denom := in[len(in)-1:]
	sizeStr := in[0 : len(in)-1]
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		return 0, err
	}
	if _, ok := sizeMap[denom]; !ok {
		return 0, fmt.Errorf("Could not find denom %s", denom)
	}
	size = size * sizeMap[denom]
	return size / G, nil
}

func generateMac() (net.HardwareAddr, error) {
	buf := make([]byte, 6)
	var mac net.HardwareAddr
	_, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}
	// Unset the local bit
	buf[0] <<= 1

	mac = append(mac, buf[0], buf[1], buf[2], buf[3], buf[4], buf[5])
	return mac, nil
}
