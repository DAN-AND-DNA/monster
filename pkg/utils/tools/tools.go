package tools

import (
	"hash/maphash"
	"math/rand"
	"strconv"
)

func FindStr(list []string, where string) bool {
	if len(list) == 0 {
		return false
	}

	for _, val := range list {
		if val == where {
			return true
		}
	}

	return false
}

func FindInt(list []int, where int) bool {
	if len(list) == 0 {
		return false
	}

	for _, val := range list {
		if val == where {
			return true
		}
	}

	return false
}

func EraseInt(list []int, index int) []int {
	if index < 0 {
		return list
	}

	if index >= len(list) {
		return list
	}

	// 删除最后一个
	if index == len(list)-1 {
		old := list[0 : len(list)-1]
		list = make([]int, len(old))

		for i, val := range old {
			list[i] = val
		}
		return list
	}

	old := list
	list = make([]int, len(old)-1)

	// 跳过中间某个
	i := 0
	for _, val := range old[:index] {
		list[i] = val
		i++
	}

	old = old[index+1:]
	for _, val := range old {
		list[i] = val
		i++
	}

	return list
}

func EraseStr(list []string, index int) []string {
	if index < 0 {
		return list
	}

	if index >= len(list) {
		return list
	}

	// 删除最后一个
	if index == len(list)-1 {
		old := list[0 : len(list)-1]
		list = make([]string, len(old))

		for i, val := range old {
			list[i] = val
		}
		return list
	}

	old := list
	list = make([]string, len(old)-1)

	// 跳过中间某个
	i := 0
	for _, val := range old[:index] {
		list[i] = val
		i++
	}

	old = old[index+1:]
	for _, val := range old {
		list[i] = val
		i++
	}

	return list
}

func HashString(s string) uint64 {
	var h maphash.Hash
	h.WriteString(s)
	return h.Sum64()
}

func HashBytes(s []byte) uint64 {
	var h maphash.Hash
	h.Write(s)
	return h.Sum64()
}

func PercentChance(percent int) bool {
	return rand.Intn(100)%100 < percent
}

func Signum(val int) int {
	if val > 0 {
		return 1
	}

	if val < 0 {
		return -1
	}

	return 0
}

func RandBetween(minVal, maxVal int) int {
	if minVal == maxVal {
		return minVal
	}

	d := maxVal - minVal
	return minVal + (rand.Int() % (d + Signum(d)))
}

func AbbreviatedKilo(amount int) string {
	if amount < 1000 {
		return strconv.Itoa(amount)
	}
	amount /= 1000
	return strconv.Itoa(amount) + "k"
}
