package detectlang

import (
	"github.com/pemistahl/lingua-go"
	"golang.org/x/text/language"
)

var tagToLingua = map[language.Tag]lingua.Language{
	language.Afrikaans:          lingua.Afrikaans,
	language.Albanian:           lingua.Albanian,
	language.Arabic:             lingua.Arabic,
	language.Armenian:           lingua.Armenian,
	language.Azerbaijani:        lingua.Azerbaijani,
	language.Bengali:            lingua.Bengali,
	language.Bulgarian:          lingua.Bulgarian,
	language.Catalan:            lingua.Catalan,
	language.SimplifiedChinese:  lingua.Chinese,
	language.TraditionalChinese: lingua.Chinese,
	language.Croatian:           lingua.Croatian,
	language.Czech:              lingua.Czech,
	language.Danish:             lingua.Danish,
	language.Dutch:              lingua.Dutch,
	language.English:            lingua.English,
	language.Estonian:           lingua.Estonian,
	language.Filipino:           lingua.Tagalog, // Filipino 映射到 lingua 的 Tagalog
	language.Finnish:            lingua.Finnish,
	language.French:             lingua.French,
	language.German:             lingua.German,
	language.Greek:              lingua.Greek,
	language.Gujarati:           lingua.Gujarati,
	language.Hebrew:             lingua.Hebrew,
	language.Hindi:              lingua.Hindi,
	language.Hungarian:          lingua.Hungarian,
	language.Icelandic:          lingua.Icelandic,
	language.Indonesian:         lingua.Indonesian,
	language.Italian:            lingua.Italian,
	language.Japanese:           lingua.Japanese,
	language.Kazakh:             lingua.Kazakh,
	language.Korean:             lingua.Korean,
	language.Latvian:            lingua.Latvian,
	language.Lithuanian:         lingua.Lithuanian,
	language.Macedonian:         lingua.Macedonian,
	language.Malay:              lingua.Malay,
	language.Marathi:            lingua.Marathi,
	language.Mongolian:          lingua.Mongolian,
	language.Norwegian:          lingua.Bokmal,
	language.Persian:            lingua.Persian,
	language.Polish:             lingua.Polish,
	language.Portuguese:         lingua.Portuguese,
	language.Punjabi:            lingua.Punjabi,
	language.Romanian:           lingua.Romanian,
	language.Russian:            lingua.Russian,
	language.Serbian:            lingua.Serbian,
	language.Slovak:             lingua.Slovak,
	language.Slovenian:          lingua.Slovene,
	language.Spanish:            lingua.Spanish,
	language.Swahili:            lingua.Swahili,
	language.Swedish:            lingua.Swedish,
	language.Tamil:              lingua.Tamil,
	language.Telugu:             lingua.Telugu,
	language.Thai:               lingua.Thai,
	language.Turkish:            lingua.Turkish,
	language.Ukrainian:          lingua.Ukrainian,
	language.Urdu:               lingua.Urdu,
	language.Vietnamese:         lingua.Vietnamese,
	language.Zulu:               lingua.Zulu,

	// Using language.Make for languages not directly in language.Tag
	language.Make("eu"): lingua.Basque,
	language.Make("be"): lingua.Belarusian,
	language.Make("nn"): lingua.Nynorsk,
	language.Make("ga"): lingua.Irish,
	language.Make("la"): lingua.Latin,
	language.Make("mi"): lingua.Maori,
	language.Make("sn"): lingua.Shona,
	language.Make("so"): lingua.Somali,
	language.Make("st"): lingua.Sotho,
	language.Make("ts"): lingua.Tsonga,
	language.Make("tn"): lingua.Tswana,
	language.Make("cy"): lingua.Welsh,
	language.Make("xh"): lingua.Xhosa,
	language.Make("yo"): lingua.Yoruba,
	language.Make("eo"): lingua.Esperanto,
	language.Make("lg"): lingua.Ganda,
}

// 创建反向映射
var linguaToTag = make(map[lingua.Language]language.Tag)

// 辅助函数：检查语言是否受支持
func IsSupportedLanguage(tag language.Tag) bool {
	_, ok := tagToLingua[tag]
	return ok
}

// 辅助函数：获取所有支持的语言标签
func GetSupportedLanguageTags() []language.Tag {
	tags := make([]language.Tag, 0, len(tagToLingua))
	for tag := range tagToLingua {
		tags = append(tags, tag)
	}
	return tags
}
