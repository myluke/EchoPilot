package jieba

import (
	"fmt"
	"testing"
)

func TestExtract(t *testing.T) {
	content := `手机卡🔴手机卡🔴手机卡🟠手机卡🟠🔴🟣🔵`
	keys := New().Tag(content)

	countMap := map[string]int{}
	for _, v := range keys {
		if v[len(v)-2:] == "/n" {
			key := v[:len(v)-2]
			fmt.Println(key)
			if _, ok := countMap[key]; ok {
				countMap[key]++
			} else {
				countMap[key] = 1
			}
		}
	}
	for _, v := range countMap {
		if v > 2 {
			fmt.Println("duiqi")
		}
	}

	fmt.Println(countMap)
}

func TestCheckRemoteDict(t *testing.T) {
	New().CheckRemoteDict()
}
