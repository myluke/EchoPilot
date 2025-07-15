package detectlang

import (
	"fmt"
	"testing"

	"golang.org/x/text/language"
)

func TestDetectLanguage(t *testing.T) {
	texts := []string{
		"这是一个简体中文的例子。",
		"這是一個繁體中文的例子。",
		"This is an English example.",
		"这是一个混合了簡體和繁體的例子。",
		"TG搜群神器 @ShopDadBot TG开店神器，自动对接@wallet 和TG Star付款！\t\tTG必备神器，发现您感兴趣的群组和频道！",
	}

	languages := []language.Tag{
		language.English,
		language.SimplifiedChinese,
		language.TraditionalChinese,
		language.Spanish,
		language.Arabic,
		language.German,
		language.Portuguese,
		language.Dutch,
		language.Polish,
		language.Danish,
		language.Russian,
		language.French,
		language.Filipino,
		language.Finnish,
		language.Korean,
		language.Czech,
		language.Swahili,
		language.Romanian,
		language.Malay,
		language.Norwegian,
		language.Japanese,
		language.Swedish,
		language.Italian,
		language.Ukrainian,
		language.Greek,
		language.Thai,
		language.Turkish,
		language.Vietnamese,
		language.Hungarian,
		language.Hebrew,
		language.Hindi,
		language.Indonesian,
	}

	client := New(languages)

	for _, text := range texts {
		result, err := client.Detect(text)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("Text: %s\n", text)
		fmt.Printf("Detected Language: %s\n", result.Language)
		fmt.Printf("Confidence: %.2f\n", result.Confidence)
		fmt.Println("---")
	}
}
