package Util

import "golang.org/x/crypto/ssh"

func GetMapKeys(m map[int]*ssh.Session) []int{
	keys := make([]int, len(m))
	for key := range m{
		keys = append(keys, key)
	}
	return keys
}
