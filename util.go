package main

import "hash/fnv"

func hash(s string) uint64 {
	algo := fnv.New64a()
	algo.Write([]byte(s))
	return algo.Sum64()
}
