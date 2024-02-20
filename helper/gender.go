package helper

import (
	"embed"
	_ "embed"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
)

//go:embed names/*
var embeddedFS embed.FS

const (
	female = "F"
	male   = "M"
)

type nameInfo struct {
	count  int
	female int
	male   int
}

// DetectGenderFromDict
func DetectGenderFromDict(name string) (string, error) {
	name = strings.ToLower(name)
	// Load data from embedded files
	files, err := embeddedFS.ReadDir("names")
	if err != nil {
		return "", err
	}

	data := make(map[string]nameInfo)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		content, err := embeddedFS.ReadFile("names/" + file.Name())
		if err != nil {
			return "", err
		}

		reader := csv.NewReader(strings.NewReader(string(content)))
		reader.Comma = ','

		rows, err := reader.ReadAll()
		if err != nil {
			return "", err
		}

		for _, row := range rows[1:] {
			if len(row) < 3 {
				continue
			}

			name := strings.ToLower(row[0])
			gender := row[1]
			count, err := strconv.Atoi(row[2])
			if err != nil {
				return "", err
			}

			info, ok := data[name]
			if !ok {
				info = nameInfo{count: count}
			} else {
				info.count += count
			}

			if gender == female {
				info.female += count
			} else if gender == male {
				info.male += count
			}

			data[name] = info
		}
	}

	// Check the gender of the given name
	info, ok := data[name]
	if !ok {
		return "", fmt.Errorf("Name not found in data set")
	}

	femaleRatio := float64(info.female) / float64(info.count)
	maleRatio := float64(info.male) / float64(info.count)

	if femaleRatio > maleRatio {
		return female, nil
	} else if maleRatio > femaleRatio {
		return male, nil
	} else {
		return "unisex", fmt.Errorf("Cannot determine gender for name")
	}
}

// 检测一个英文名的性别
func DetermineGender(name string) string {
	name = strings.ToLower(name)

	// 处理一些特殊情况
	switch name {
	case "lee", "jamie", "jordan", "kelly":
		return "unisex"
	case "pat", "jody":
		return "unisex"
	}

	// 处理常见词缀
	if strings.HasPrefix(name, "chr") || strings.HasPrefix(name, "cath") {
		return "female"
	} else if strings.HasPrefix(name, "alex") || strings.HasPrefix(name, "max") {
		return "male"
	}

	// 判断名字中的元音字母和辅音字母
	vowelCount := 0
	consonantCount := 0
	for _, c := range name {
		if strings.ContainsAny(string(c), "aeiouy") {
			vowelCount++
		} else {
			consonantCount++
		}
	}

	// 处理以辅音字母+ie结尾的名称
	if strings.HasSuffix(name, "ie") && !strings.ContainsAny(string(name[len(name)-3]), "aeiouy") {
		return "female"
	}

	// 处理以y结尾的名称
	lastLetter := string(name[len(name)-1])
	if lastLetter == "y" {
		if strings.ContainsAny(string(name[len(name)-2]), "aeiouy") {
			return "unisex"
		} else {
			return "male"
		}
	}

	// 处理以son结尾的名称
	if strings.HasSuffix(name, "son") {
		return "male"
	}

	// 处理以berg结尾的名称
	if strings.HasSuffix(name, "berg") {
		return "male"
	}

	// 处理以li结尾的名称
	if strings.HasSuffix(name, "li") {
		return "male"
	}

	// 处理以ell结尾的名称
	if strings.HasSuffix(name, "ell") {
		return "male"
	}

	// 处理以ette结尾的名称
	if strings.HasSuffix(name, "ette") || strings.HasSuffix(name, "elle") || strings.HasSuffix(name, "ine") {
		return "female"
	}

	// 处理以ard结尾的名称
	if strings.HasSuffix(name, "ard") || strings.HasSuffix(name, "old") {
		return "male"
	}

	// 处理以drew结尾的名称
	if strings.HasSuffix(name, "drew") {
		return "unisex"
	}

	// 处理以nny结尾的名称
	if strings.HasSuffix(name, "nny") || strings.HasSuffix(name, "ria") {
		return "female"
	}

	// 处理以lou结尾的名称
	if strings.HasSuffix(name, "lou") {
		return "unisex"
	}

	// 处理以er结尾的名称
	if strings.HasSuffix(name, "er") {
		if vowelCount == 1 && consonantCount == 3 {
			return "male"
		} else if vowelCount == 2 && consonantCount == 3 {
			return "unisex"
		} else if vowelCount > consonantCount {
			return "female"
		} else if consonantCount > vowelCount {
			return "male"
		}
	}

	// 根据元音字母和辅音字母的数量来判断性别
	if vowelCount == 1 && consonantCount == 2 {
		return "male"
	} else if vowelCount == 2 && consonantCount == 2 {
		return "unisex"
	} else if vowelCount > consonantCount {
		return "female"
	} else if consonantCount > vowelCount {
		return "male"
	}

	// 如果以上规则无法确定性别，则返回“未知”
	return "unknown"
}
