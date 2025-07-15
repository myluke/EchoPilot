package sensitive

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestSensitiveFilter(t *testing.T) {
	dictURL := "https://raw.githubusercontent.com/mylukin/sensitive-service/main/dict.txt"

	filter, err := Get(
		dictURL,
		WithUpdateInterval(10*time.Second),
		WithHTTPClient(&http.Client{Timeout: 15 * time.Second}),
		WithDebugMode(true),
	)
	if err != nil {
		t.Fatalf("初始化敏感词过滤器失败: %v", err)
	}
	defer filter.Close()
	filter.AddWord("垃圾", "傻逼", "混蛋", "白痴", "笨蛋", "废物")
	// 打印当前敏感词列表的一部分，用于调试
	fmt.Println("敏感词示例:")
	words := filter.FindAll("垃圾 傻逼 混蛋 白痴 笨蛋 废物")
	for _, word := range words {
		fmt.Printf("- %s\n", word)
	}

	t.Run("AddWord", func(t *testing.T) {
		filter.AddWord("测试词")
		if found, word := filter.FindIn("这是一个测试词"); !found || word != "测试词" {
			t.Errorf("AddWord() 失败，期望找到 '测试词'，实际结果: found=%v, word=%s", found, word)
		}
	})

	t.Run("Replace", func(t *testing.T) {
		text := "这篇文章真的好垃圾"
		replacedText := filter.Replace(text, '*')
		expected := "这篇文章真的好**"
		if replacedText != expected {
			t.Errorf("Replace() = %s, 期望 %s", replacedText, expected)
		}
	})

	t.Run("Filter", func(t *testing.T) {
		text := "这篇文章真的好垃圾啊"
		filteredText := filter.Filter(text)
		expected := "这篇文章真的好啊"
		if filteredText != expected {
			t.Errorf("Filter() = %s, 期望 %s", filteredText, expected)
		}
	})

	t.Run("FindIn", func(t *testing.T) {
		text := "这是一个测试语句，包含垃圾这个词"
		found, word := filter.FindIn(text)
		if !found {
			t.Errorf("FindIn() 应该找到敏感词")
		}
		if word != "垃圾" {
			t.Errorf("FindIn() 找到的词 = %s, 期望 '垃圾'", word)
		}
	})

	t.Run("Validate", func(t *testing.T) {
		text := "这是一个正常的句子"
		valid, _ := filter.Validate(text)
		if !valid {
			t.Errorf("Validate() 应该返回 true 对于正常句子")
		}

		text = "这句话包含垃圾这个词"
		valid, invalidWord := filter.Validate(text)
		if valid {
			t.Errorf("Validate() 应该返回 false 对于包含敏感词的句子")
		}
		if invalidWord != "垃圾" {
			t.Errorf("Validate() 返回的无效词 = %s, 期望 '垃圾'", invalidWord)
		}
	})

	t.Run("FindAll", func(t *testing.T) {
		text := "这句话包含多个敏感词，如垃圾和废物"
		found := filter.FindAll(text)
		expected := []string{"垃圾", "废物"}
		if len(found) != len(expected) {
			t.Errorf("FindAll() 找到 %d 个词, 期望 %d 个", len(found), len(expected))
		}
		for i, word := range found {
			if word != expected[i] {
				t.Errorf("FindAll() 找到的词 %s, 期望 %s", word, expected[i])
			}
		}
	})

	t.Run("UpdateNoisePattern", func(t *testing.T) {
		filter.UpdateNoisePattern(`x`)
		found, word := filter.FindIn("这篇文章真的好垃x圾")
		if !found || word != "垃圾" {
			t.Errorf("UpdateNoisePattern() 后 FindIn() 失败，期望找到 '垃圾'，实际结果: found=%v, word=%s", found, word)
		}
	})

	t.Run("Length", func(t *testing.T) {
		length := filter.Length()
		if length == 0 {
			t.Errorf("Length() 返回 0, 期望大于 0")
		}
		fmt.Printf("当前敏感词数量: %d\n", length)
	})

	time.Sleep(10 * time.Minute)
}
