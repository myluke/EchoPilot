package helper

import "strings"

// GetLanguageCode 获取国家代码对应的主要语言代码
func GetLanguageCode(countryCode string) string {
	countryCode = strings.ToUpper(countryCode)
	var countryToLanguage = map[string]string{
		"AD": "ca", "AE": "ar-AE", "AF": "ps", "AG": "en-AG", "AI": "en-AI", "AL": "sq", "AM": "hy", "AO": "pt-AO",
		"AR": "es-AR", "AS": "en-AS", "AT": "de-AT", "AU": "en-AU", "AW": "nl-AW", "AX": "sv-AX", "AZ": "az",
		"BA": "bs", "BB": "en-BB", "BD": "bn", "BE": "nl-BE", "BF": "fr-BF", "BG": "bg", "BH": "ar-BH", "BI": "rn",
		"BJ": "fr-BJ", "BL": "fr-BL", "BM": "en-BM", "BN": "ms-BN", "BO": "es-BO", "BQ": "nl-BQ", "BR": "pt-BR",
		"BS": "en-BS", "BT": "dz", "BV": "no", "BW": "en-BW", "BY": "be", "BZ": "en-BZ", "CA": "en-CA",
		"CC": "en-CC", "CD": "fr-CD", "CF": "fr-CF", "CG": "fr-CG", "CH": "de-CH", "CI": "fr-CI", "CK": "en-CK",
		"CL": "es-CL", "CM": "en-CM", "CN": "zh-Hans", "CO": "es-CO", "CR": "es-CR", "CU": "es-CU", "CV": "pt-CV",
		"CW": "nl-CW", "CX": "en-CX", "CY": "el-CY", "CZ": "cs", "DE": "de-DE", "DJ": "fr-DJ", "DK": "da",
		"DM": "en-DM", "DO": "es-DO", "DZ": "ar-DZ", "EC": "es-EC", "EE": "et", "EG": "ar-EG", "EH": "ar-EH",
		"ER": "ti", "ES": "es-ES", "ET": "am", "FI": "fi", "FJ": "en-FJ", "FK": "en-FK", "FM": "en-FM", "FO": "fo",
		"FR": "fr-FR", "GA": "fr-GA", "GB": "en-GB", "GD": "en-GD", "GE": "ka", "GF": "fr-GF", "GG": "en-GG",
		"GH": "en-GH", "GI": "en-GI", "GL": "kl", "GM": "en-GM", "GN": "fr-GN", "GP": "fr-GP", "GQ": "es-GQ",
		"GR": "el-GR", "GS": "en-GS", "GT": "es-GT", "GU": "en-GU", "GW": "pt-GW", "GY": "en-GY", "HK": "zh-Hant",
		"HM": "en-HM", "HN": "es-HN", "HR": "hr", "HT": "ht", "HU": "hu", "ID": "id", "IE": "en-IE", "IL": "he",
		"IM": "en-IM", "IN": "hi", "IO": "en-IO", "IQ": "ar-IQ", "IR": "fa", "IS": "is", "IT": "it-IT", "JE": "en-JE",
		"JM": "en-JM", "JO": "ar-JO", "JP": "ja", "KE": "sw-KE", "KG": "ky", "KH": "km", "KI": "en-KI", "KM": "ar-KM",
		"KN": "en-KN", "KP": "ko-KP", "KR": "ko-KR", "KW": "ar-KW", "KY": "en-KY", "KZ": "kk", "LA": "lo",
		"LB": "ar-LB", "LC": "en-LC", "LI": "de-LI", "LK": "si", "LR": "en-LR", "LS": "en-LS", "LT": "lt",
		"LU": "fr-LU", "LV": "lv", "LY": "ar-LY", "MA": "ar-MA", "MC": "fr-MC", "MD": "ro-MD", "ME": "sr-ME",
		"MF": "fr-MF", "MG": "mg", "MH": "en-MH", "MK": "mk", "ML": "fr-ML", "MM": "my", "MN": "mn", "MO": "zh-Hant",
		"MP": "en-MP", "MQ": "fr-MQ", "MR": "ar-MR", "MS": "en-MS", "MT": "mt", "MU": "en-MU", "MV": "dv",
		"MW": "en-MW", "MX": "es-MX", "MY": "ms-MY", "MZ": "pt-MZ", "NA": "en-NA", "NC": "fr-NC", "NE": "fr-NE",
		"NF": "en-NF", "NG": "en-NG", "NI": "es-NI", "NL": "nl-NL", "NO": "nb", "NP": "ne", "NR": "en-NR",
		"NU": "en-NU", "NZ": "en-NZ", "OM": "ar-OM", "PA": "es-PA", "PE": "es-PE", "PF": "fr-PF", "PG": "en-PG",
		"PH": "en-PH", "PK": "ur", "PL": "pl", "PM": "fr-PM", "PN": "en-PN", "PR": "es-PR", "PS": "ar-PS",
		"PT": "pt-PT", "PW": "en-PW", "PY": "es-PY", "QA": "ar-QA", "RE": "fr-RE", "RO": "ro", "RS": "sr", "RU": "ru",
		"RW": "rw", "SA": "ar-SA", "SB": "en-SB", "SC": "en-SC", "SD": "ar-SD", "SE": "sv-SE", "SG": "en-SG",
		"SH": "en-SH", "SI": "sl", "SJ": "nb-SJ", "SK": "sk", "SL": "en-SL", "SM": "it-SM", "SN": "fr-SN", "SO": "so",
		"SR": "nl-SR", "SS": "en-SS", "ST": "pt-ST", "SV": "es-SV", "SX": "nl-SX", "SY": "ar-SY", "SZ": "en-SZ",
		"TC": "en-TC", "TD": "ar-TD", "TF": "fr-TF", "TG": "fr-TG", "TH": "th", "TJ": "tg", "TK": "en-TK",
		"TL": "pt-TL", "TM": "tk", "TN": "ar-TN", "TO": "en-TO", "TR": "tr", "TT": "en-TT", "TV": "en-TV",
		"TW": "zh-Hant", "TZ": "sw-TZ", "UA": "uk", "UG": "en-UG", "UM": "en-UM", "US": "en-US", "UY": "es-UY",
		"UZ": "uz", "VA": "it", "VC": "en-VC", "VE": "es-VE", "VG": "en-VG", "VI": "en-VI", "VN": "vi",
		"VU": "bi", "WF": "fr-WF", "WS": "sm", "YE": "ar-YE", "YT": "fr-YT", "ZA": "en-ZA", "ZM": "en-ZM", "ZW": "en-ZW",
	}
	if lang, ok := countryToLanguage[countryCode]; ok {
		return strings.ToLower(lang)
	}
	return ""
}
