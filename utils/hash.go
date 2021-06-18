package utils

import "hash/fnv"

func Hash(s string) uint64 {
	algo := fnv.New64a()
	algo.Write([]byte(s))
	return algo.Sum64()
}
