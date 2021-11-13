package swarm

import (
	"encoding/binary"
	"encoding/hex"
	"math"
)

const invalidSwarmID = uint64(math.MaxUint64)
const maxSwarmID = invalidSwarmID - 1

func hexdigit(d byte) (h uint64) {
	if d >= 48 && d <= 57 {
		h = uint64(d - 48)
	}
	if d >= 65 && d <= 70 {
		h = uint64(d - 55)
	}
	if d >= 97 && d <= 102 {
		h = uint64(d - 87)
	}
	return
}

func pubkeyToUInt64(pk string) (res uint64) {
	data, _ := hex.DecodeString(pk)
	for i := 0; i < len(data); i += 8 {
		res ^= binary.BigEndian.Uint64(data[i : i+8])
	}
	return
}

/// GetSwarmForPubkey gets the swarm the public key belongs in
func GetSwarmForPubkey(snodes []ServiceNode, pk string) (swarm []ServiceNode) {
	res := pubkeyToUInt64(pk)
	id := invalidSwarmID
	mindist := uint64(math.MaxUint64)

	for _, node := range snodes {
		dist := uint64(0)
		if res > node.SwarmID {
			dist = res - node.SwarmID
		} else {
			dist = node.SwarmID - res
		}
		if dist < mindist {
			mindist = dist
			id = node.SwarmID
		}
	}
	for _, node := range snodes {
		if node.SwarmID == id {
			swarm = append(swarm, node)
		}
	}
	return
}
