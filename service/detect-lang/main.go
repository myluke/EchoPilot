package detectlang

import (
	"fmt"

	"github.com/labstack/gommon/log"
	"github.com/longbridgeapp/opencc"
	"github.com/pemistahl/lingua-go"
	"golang.org/x/text/language"
)

// LanguageDetectionResult stores the result of language detection
type LanguageDetectionResult struct {
	Language   language.Tag
	Confidence float64
}

var s2tConverter *opencc.OpenCC
var t2sConverter *opencc.OpenCC

func init() {
	var err error
	s2tConverter, err = opencc.New("s2t")
	if err != nil {
		panic(err)
	}
	t2sConverter, err = opencc.New("t2s")
	if err != nil {
		panic(err)
	}

	for tag, v := range tagToLingua {
		// For Chinese, we choose to use SimplifiedChinese as the default mapping
		if v == lingua.Chinese && linguaToTag[v] == language.Und {
			linguaToTag[v] = language.SimplifiedChinese
		} else {
			linguaToTag[v] = tag
		}
	}
}

type Client struct {
	detector lingua.LanguageDetector
	langs    []lingua.Language
}

// New initialization function
func New(languages []language.Tag) *Client {
	var linguaLanguages []lingua.Language
	for _, tag := range languages {
		if linguaLang, ok := tagToLingua[tag]; ok {
			linguaLanguages = append(linguaLanguages, linguaLang)
		} else {
			log.Warnf("unsupported language: %v", tag)
		}
	}

	detector := lingua.NewLanguageDetectorBuilder().
		FromLanguages(linguaLanguages...).
		Build()

	return &Client{
		detector: detector,
		langs:    linguaLanguages,
	}
}

// Detect Chinese variant (Simplified or Traditional)
func detectChineseVariant(text string) language.Tag {
	// 只处理前 1000 个字符
	maxLen := 1000
	if len(text) > maxLen {
		text = text[:maxLen]
	}

	s2tText, err := s2tConverter.Convert(text)
	if err != nil {
		return language.SimplifiedChinese
	}

	diffCount := 0
	for i := 0; i < len(text); i++ {
		if text[i] != s2tText[i] {
			diffCount++
		}
	}

	// 如果差异字符数量超过 10%，认为是简体中文
	if float64(diffCount)/float64(len(text)) > 0.1 {
		return language.SimplifiedChinese
	}
	return language.TraditionalChinese
}

// Calculate Levenshtein distance
func levenshteinDistance(s, t string) int {
	d := make([][]int, len(s)+1)
	for i := range d {
		d[i] = make([]int, len(t)+1)
	}
	for i := range d {
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}
	for j := 1; j <= len(t); j++ {
		for i := 1; i <= len(s); i++ {
			if s[i-1] == t[j-1] {
				d[i][j] = d[i-1][j-1]
			} else {
				min := d[i-1][j]
				if d[i][j-1] < min {
					min = d[i][j-1]
				}
				if d[i-1][j-1] < min {
					min = d[i-1][j-1]
				}
				d[i][j] = min + 1
			}
		}
	}
	return d[len(s)][len(t)]
}

// DetectLanguage detects the language of the given text
func (c *Client) Detect(text string) (*LanguageDetectionResult, error) {
	if text == "" {
		return nil, fmt.Errorf("input text is empty")
	}
	if c.detector == nil {
		return nil, fmt.Errorf("language detector is not initialized, please call InitializeDetector first")
	}

	detectedLang, exists := c.detector.DetectLanguageOf(text)
	if !exists {
		return nil, fmt.Errorf("unable to detect language")
	}

	confidence := c.detector.ComputeLanguageConfidence(text, detectedLang)

	// Convert the result back to language.Tag
	resultTag, ok := linguaToTag[detectedLang]
	if !ok {
		return nil, fmt.Errorf("unable to map detected language to language.Tag")
	}

	// If Chinese is detected, further determine Simplified or Traditional
	if resultTag == language.SimplifiedChinese || resultTag == language.TraditionalChinese {
		resultTag = detectChineseVariant(text)
	}

	return &LanguageDetectionResult{
		Language:   resultTag,
		Confidence: confidence,
	}, nil
}
