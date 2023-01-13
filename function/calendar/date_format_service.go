package calendar

type DateFormatLocaleService struct {
}

func (DateFormatLocaleService) GetLocaleByTag(tag string) Locale {
	locales := listLocales()
	for _, l := range locales {
		if string(l) == tag {
			return l
		}
	}
	return LocaleEnUS

}

func (DateFormatLocaleService) GetFullFormatsByLocale(locale Locale) string {
	format, ok := fullFormatsByLocale[locale]
	if !ok {
		return DefaultFormatEnUSFull
	}

	return format
}

func (DateFormatLocaleService) GetLongFormatsByLocale(locale Locale) string {

	format, ok := longFormatsByLocale[locale]
	if !ok {
		return DefaultFormatEnUSLong
	}

	return format

}

func (DateFormatLocaleService) GetMediumFormatsByLocale(locale Locale) string {
	format, ok := mediumFormatsByLocale[locale]
	if !ok {
		return DefaultFormatEnUSMedium
	}

	return format
}

func (DateFormatLocaleService) GetShortFormatsByLocale(locale Locale) string {
	format, ok := shortFormatsByLocale[locale]
	if !ok {
		return DefaultFormatEnUSShort
	}

	return format
}

func (DateFormatLocaleService) GetDateTimeFormatsByLocale(locale Locale) string {

	format, ok := dateTimeFormatsByLocale[locale]
	if !ok {
		return DefaultFormatEnUSDateTime
	}

	return format
}

func (DateFormatLocaleService) GetTimeFormatsByLocale(locale Locale) string {

	format, ok := TimeFormatsByLocale[locale]
	if !ok {
		return DefaultFormatEnUSTime
	}

	return format
}

type Locale string

const (
	LocaleEn   = "en"
	LocaleDe   = "de"
	LocaleEs   = "es"
	LocaleFr   = "fr"
	LocaleIt   = "it"
	LocaleHu   = "hu"
	LocaleNl   = "nl"
	LocalePl   = "pl"
	LocaleRo   = "ro"
	LocaleSv   = "sv"
	LocaleTr   = "tr"
	LocaleBg   = "bg"
	LocaleRu   = "ru"
	LocaleFa   = "fa"
	LocaleKo   = "ko"
	LocaleJa   = "ja"
	LocaleUk   = "uk"
	LocaleEnAU = "en-AU"
	LocalePtBR = "pt-BR" // Portuguese (Brazil)
	LocaleZhCN = "zh-CN" // Chinese (Mainland)
	LocaleZhTW = "zh-TW" // Chinese (Taiwan)
	LocaleEnUS = "en-US" // English (United States)
	LocaleEnGB = "en-GB" // English (United Kingdom)
	LocaleDaDK = "da-DK" // Danish (Denmark)
	LocaleNlBE = "nl-BE" // Dutch (Belgium)
	LocaleNlNL = "nl-NL" // Dutch (Netherlands)
	LocaleFiFI = "fi-FI" // Finnish (Finland)
	LocaleFrFR = "fr-FR" // French (France)
	LocaleFrCA = "fr-CA" // French (Canada)
	LocaleDeDE = "de-DE" // German (Germany)
	LocaleHuHU = "hu-HU" // Hungarian (Hungary)
	LocaleItIT = "it-IT" // Italian (Italy)
	LocaleNnNO = "nn-NO" // Norwegian Nynorsk (Norway)
	LocaleNbNO = "nb-NO" // Norwegian Bokmål (Norway)
	LocalePlPL = "pl-PL" // Polish (Poland)
	LocalePtPT = "pt-PT" // Portuguese (Portugal)
	LocaleRoRO = "ro-RO" // Romanian (Romania)
	LocaleRuRU = "ru-RU" // Russian (Russia)
	LocaleEsES = "es-ES" // Spanish (Spain)
	LocaleCaES = "ca-ES" // Catalan (Spain)
	LocaleSvSE = "sv-SE" // Swedish (Sweden)
	LocaleTrTR = "tr-TR" // Turkish (Turkey)
	LocaleUkUA = "uk-UA" // Ukrainian (Ukraine)
	LocaleBgBG = "bg-BG" // Bulgarian (Bulgaria)
	LocaleZhHK = "zh-HK" // Chinese (Hong Kong)
	LocaleKoKR = "ko-KR" // Korean (Korea)
	LocaleJaJP = "ja-JP" // Japanese (Japan)
	LocaleElGR = "el-GR" // Greek (Greece)
	LocaleFrGP = "fr-GP" // French (Guadeloupe)
	LocaleFrLU = "fr-LU" // French (Luxembourg)
	LocaleFrMQ = "fr-MQ" // French (Martinique)
	LocaleFrRE = "fr-RE" // French (Reunion)
	LocaleFrGF = "fr-GF" // French (French Guiana)
	LocaleCsCZ = "cs-CZ" // Czech (Czech Republic)
	LocaleSlSI = "sl-SI" // Slovenian (Slovenia)
	LocaleLtLT = "lt-LT" // Lithuanian (Lithuania)
	LocaleThTH = "th-TH" // Thai (Thailand)
	LocaleUzUZ = "uz-UZ" // Uzbek (Uzbekistan)
)

// ListLocales returns all locales supported by the package.
func listLocales() []Locale {
	return []Locale{
		LocaleEn,
		LocaleDe,
		LocaleEs,
		LocaleFr,
		LocaleIt,
		LocaleHu,
		LocaleNl,
		LocalePl,
		LocaleRo,
		LocaleSv,
		LocaleTr,
		LocaleBg,
		LocaleRu,
		LocaleFa,
		LocaleKo,
		LocaleJa,
		LocaleUk,
		LocaleEnAU,

		LocaleEnUS,
		LocaleEnGB,
		LocaleDaDK,
		LocaleNlBE,
		LocaleNlNL,
		LocaleFiFI,
		LocaleFrFR,
		LocaleFrCA,
		LocaleDeDE,
		LocaleHuHU,
		LocaleItIT,
		LocaleNnNO,
		LocaleNbNO,
		LocalePlPL,
		LocalePtPT,
		LocalePtBR,
		LocaleRoRO,
		LocaleRuRU,
		LocaleEsES,
		LocaleCaES,
		LocaleSvSE,
		LocaleTrTR,
		LocaleUkUA,
		LocaleBgBG,
		LocaleZhCN,
		LocaleZhTW,
		LocaleZhHK,
		LocaleKoKR,
		LocaleJaJP,
		LocaleElGR,
		LocaleFrGP,
		LocaleFrLU,
		LocaleFrMQ,
		LocaleFrRE,
		LocaleFrGF,
		LocaleCsCZ,
		LocaleSlSI,
		LocaleLtLT,
		LocaleThTH,
		LocaleUzUZ,
	}
}

const (
	DefaultFormatEnUSFull     = "Monday, January 2" // English (United States)
	DefaultFormatEnUSLong     = "January 2, 2006"
	DefaultFormatEnUSMedium   = "Jan 02, 2006"
	DefaultFormatEnUSShort    = "1/2/06"
	DefaultFormatEnUSDateTime = "1/2/06 3:04 PM"
	DefaultFormatEnUSTime     = "3:04 PM"

	DefaultFormatEnGBFull     = "Monday, 2 January" // English (United Kingdom)
	DefaultFormatEnGBLong     = "2 January 2006"
	DefaultFormatEnGBMedium   = "02 Jan 2006"
	DefaultFormatEnGBShort    = "02/01/2006"
	DefaultFormatEnGBDateTime = "02/01/2006 15:04"
	DefaultFormatEnGBTime     = "15:04"

	DefaultFormatDaDKFull     = "Monday den 2. January" // Danish (Denmark)
	DefaultFormatDaDKLong     = "2. Jan 2006"
	DefaultFormatDaDKMedium   = "02/01/2006"
	DefaultFormatDaDKShort    = "02/01/06"
	DefaultFormatDaDKDateTime = "02/01/2006 15.04"
	DefaultFormatDaDKTime     = "15.04"

	DefaultFormatNlBEFull     = "Monday 2 January" // Dutch (Belgium)
	DefaultFormatNlBELong     = "2 January 2006"
	DefaultFormatNlBEMedium   = "02-Jan-2006"
	DefaultFormatNlBEShort    = "2/01/06"
	DefaultFormatNlBEDateTime = "2/01/06 15:04"
	DefaultFormatNlBETime     = "15:04"

	DefaultFormatNlNLFull     = "Monday 2 January" // Dutch (Netherlands)
	DefaultFormatNlNLLong     = "2 January 2006"
	DefaultFormatNlNLMedium   = "02 Jan 2006"
	DefaultFormatNlNLShort    = "02-01-06"
	DefaultFormatNlNLDateTime = "02-01-06 15:04"
	DefaultFormatNlNLTime     = "15:04"

	DefaultFormatFiFIFull     = "Monday 2. January" // Finnish (Finland)
	DefaultFormatFiFILong     = "2. January 2006"
	DefaultFormatFiFIMedium   = "02.1.2006"
	DefaultFormatFiFIShort    = "02.1.2006"
	DefaultFormatFiFIDateTime = "02.1.2006 15.04"
	DefaultFormatFiFITime     = "15.04"

	DefaultFormatFrFRFull     = "Monday 2 January" // French (France)
	DefaultFormatFrFRLong     = "2 January 2006"
	DefaultFormatFrFRMedium   = "02 Jan 2006"
	DefaultFormatFrFRShort    = "02/01/2006"
	DefaultFormatFrFRDateTime = "02/01/2006 15:04"
	DefaultFormatFrFRTime     = "15:04"

	DefaultFormatFrCAFull     = "Monday 2 January" // French (Canada)
	DefaultFormatFrCALong     = "2 January 2006"
	DefaultFormatFrCAMedium   = "2006-01-02"
	DefaultFormatFrCAShort    = "06-01-02"
	DefaultFormatFrCADateTime = "06-01-02 15:04"
	DefaultFormatFrCATime     = "15:04"

	DefaultFormatFrGPFull     = "Monday 2 January" // French (Guadeloupe)
	DefaultFormatFrGPLong     = "2 January 2006"
	DefaultFormatFrGPMedium   = "2006-01-02"
	DefaultFormatFrGPShort    = "06-01-02"
	DefaultFormatFrGPDateTime = "06-01-02 15:04"
	DefaultFormatFrGPTime     = "15:04"

	DefaultFormatFrLUFull     = "Monday 2 January" // French (Luxembourg)
	DefaultFormatFrLULong     = "2 January 2006"
	DefaultFormatFrLUMedium   = "2006-01-02"
	DefaultFormatFrLUShort    = "06-01-02"
	DefaultFormatFrLUDateTime = "06-01-02 15:04"
	DefaultFormatFrLUTime     = "15:04"

	DefaultFormatFrMQFull     = "Monday 2 January" // French (Martinique)
	DefaultFormatFrMQLong     = "2 January 2006"
	DefaultFormatFrMQMedium   = "2006-01-02"
	DefaultFormatFrMQShort    = "06-01-02"
	DefaultFormatFrMQDateTime = "06-01-02 15:04"
	DefaultFormatFrMQTime     = "15:04"

	DefaultFormatFrGFFull     = "Monday 2 January" // French (French Guiana)
	DefaultFormatFrGFLong     = "2 January 2006"
	DefaultFormatFrGFMedium   = "2006-01-02"
	DefaultFormatFrGFShort    = "06-01-02"
	DefaultFormatFrGFDateTime = "06-01-02 15:04"
	DefaultFormatFrGFTime     = "15:04"

	DefaultFormatFrREFull     = "Monday 2 January" // French (Reunion)
	DefaultFormatFrRELong     = "2 January 2006"
	DefaultFormatFrREMedium   = "2006-01-02"
	DefaultFormatFrREShort    = "06-01-02"
	DefaultFormatFrREDateTime = "06-01-02 15:04"
	DefaultFormatFrRETime     = "15:04"

	DefaultFormatDeDEFull     = "Monday, 2. January" // German (Germany)
	DefaultFormatDeDELong     = "2. January 2006"
	DefaultFormatDeDEMedium   = "02.01.2006"
	DefaultFormatDeDEShort    = "02.01.06"
	DefaultFormatDeDEDateTime = "02.01.06 15:04"
	DefaultFormatDeDETime     = "15:04"

	DefaultFormatPlFull     = "Monday, 2. January" // Polish (Poland)
	DefaultFormatPlLong     = "2. January 2006"
	DefaultFormatPlMedium   = "02.01.2006"
	DefaultFormatPlShort    = "02.01.06"
	DefaultFormatPlDateTime = "02.01.06 15:04"
	DefaultFormatPlTime     = "15:04"

	DefaultFormatHuHUFull     = "January 2., Monday" // Hungarian (Hungary)
	DefaultFormatHuHULong     = "2006. January 2."
	DefaultFormatHuHUMedium   = "2006.01.02."
	DefaultFormatHuHUShort    = "2006.01.02."
	DefaultFormatHuHUDateTime = "2006.01.02. 15:04"
	DefaultFormatHuHUTime     = "15:04"

	DefaultFormatFaFull     = "January 2., Monday" // Persian
	DefaultFormatFaLong     = "2006. January 2."
	DefaultFormatFaMedium   = "2006.01.02."
	DefaultFormatFaShort    = "2006.01.02."
	DefaultFormatFaDateTime = "2006.01.02. 15:04"
	DefaultFormatFaTime     = "15:04"

	DefaultFormatItITFull     = "Monday 2 January" // Italian (Italy)
	DefaultFormatItITLong     = "2 January 2006"
	DefaultFormatItITMedium   = "02/Jan/2006"
	DefaultFormatItITShort    = "02/01/06"
	DefaultFormatItITDateTime = "02/01/06 15:04"
	DefaultFormatItITTime     = "15:04"

	DefaultFormatNnNOFull     = "Monday 2. January" // Norwegian Nynorsk (Norway)
	DefaultFormatNnNOLong     = "2. January 2006"
	DefaultFormatNnNOMedium   = "02. Jan 2006"
	DefaultFormatNnNOShort    = "02.01.06"
	DefaultFormatNnNODateTime = "02.01.06 15:04"
	DefaultFormatNnNOTime     = "15:04"

	DefaultFormatNbNOFull     = "Monday 2. January" // Norwegian Bokmål (Norway)
	DefaultFormatNbNOLong     = "2. January 2006"
	DefaultFormatNbNOMedium   = "02. Jan 2006"
	DefaultFormatNbNOShort    = "02.01.06"
	DefaultFormatNbNODateTime = "15:04 02.01.06"
	DefaultFormatNbNOTime     = "15:04"

	DefaultFormatPtPTFull     = "Monday, 2 de January" // Portuguese (Portugal)
	DefaultFormatPtPTLong     = "2 de January de 2006"
	DefaultFormatPtPTMedium   = "02/01/2006"
	DefaultFormatPtPTShort    = "02/01/06"
	DefaultFormatPtPTDateTime = "02/01/06, 15:04"
	DefaultFormatPtPTTime     = "15:04"

	DefaultFormatPtBRFull     = "Monday, 2 de January" // Portuguese (Brazil)
	DefaultFormatPtBRLong     = "02 de January de 2006"
	DefaultFormatPtBRMedium   = "02/01/2006"
	DefaultFormatPtBRShort    = "02/01/06"
	DefaultFormatPtBRDateTime = "02/01/06, 15:04"
	DefaultFormatPtBRTime     = "15:04"

	DefaultFormatRoROFull     = "Monday, 02 January" // Romanian (Romania)
	DefaultFormatRoROLong     = "02 January 2006"
	DefaultFormatRoROMedium   = "02.01.2006"
	DefaultFormatRoROShort    = "02.01.2006"
	DefaultFormatRoRODateTime = "02.01.06, 15:04"
	DefaultFormatRoROTime     = "15:04"

	DefaultFormatRuRUFull     = "Monday, 2 January" // Russian (Russia)
	DefaultFormatRuRULong     = "2 January 2006 г."
	DefaultFormatRuRUMedium   = "02 Jan 2006 г."
	DefaultFormatRuRUShort    = "02.01.06"
	DefaultFormatRuRUDateTime = "02.01.06, 15:04"
	DefaultFormatRuRUTime     = "15:04"

	DefaultFormatEsESFull     = "Monday, 2 de January" // Spanish (Spain)
	DefaultFormatEsESLong     = "2 de January de 2006"
	DefaultFormatEsESMedium   = "02/01/2006"
	DefaultFormatEsESShort    = "02/01/06"
	DefaultFormatEsESDateTime = "02/01/06 15:04"
	DefaultFormatEsESTime     = "15:04"

	DefaultFormatCaESFull     = "Monday, 2 de January" // Spanish (Spain)
	DefaultFormatCaESLong     = "2 de January de 2006"
	DefaultFormatCaESMedium   = "02/01/2006"
	DefaultFormatCaESShort    = "02/01/06"
	DefaultFormatCaESDateTime = "02/01/06 15:04"
	DefaultFormatCaESTime     = "15:04"

	DefaultFormatSvSEFull     = "Mondayen den 2:e January" // Swedish (Sweden)
	DefaultFormatSvSELong     = "2 January 2006"
	DefaultFormatSvSEMedium   = "2 Jan 2006"
	DefaultFormatSvSEShort    = "2006-01-02"
	DefaultFormatSvSEDateTime = "2006-01-02 15:04"
	DefaultFormatSvSETime     = "15:04"

	DefaultFormatTrTRFull     = "2 January Monday" // Turkish (Turkey)
	DefaultFormatTrTRLong     = "2 January 2006"
	DefaultFormatTrTRMedium   = "2 Jan 2006"
	DefaultFormatTrTRShort    = "2.01.2006"
	DefaultFormatTrTRDateTime = "2.01.2006 15:04"
	DefaultFormatTrTRTime     = "15:04"

	DefaultFormatUkUAFull     = "Monday, 2 January" // Ukrainian (Ukraine)
	DefaultFormatUkUALong     = "2 January 2006 р."
	DefaultFormatUkUAMedium   = "02 Jan 2006 р."
	DefaultFormatUkUAShort    = "02.01.06"
	DefaultFormatUkUADateTime = "02.01.06, 15:04"
	DefaultFormatUkUATime     = "15:04"

	DefaultFormatBgBGFull     = "Monday, 2 January" // Bulgarian (Bulgaria)
	DefaultFormatBgBGLong     = "2 January 2006"
	DefaultFormatBgBGMedium   = "2 Jan 2006"
	DefaultFormatBgBGShort    = "2.01.2006"
	DefaultFormatBgBGDateTime = "2.01.2006 15:04"
	DefaultFormatBgBGTime     = "15:04"

	DefaultFormatZhCNFull     = "年1月2日 Monday" // Chinese (Mainland)
	DefaultFormatZhCNLong     = "2006年1月2日"
	DefaultFormatZhCNMedium   = "2006-01-02"
	DefaultFormatZhCNShort    = "2006/1/2"
	DefaultFormatZhCNDateTime = "2006-01-02 15:04"
	DefaultFormatZhCNTime     = "15:04"

	DefaultFormatZhTWFull     = "年1月2日 Monday" // Chinese (Taiwan)
	DefaultFormatZhTWLong     = "2006年1月2日"
	DefaultFormatZhTWMedium   = "2006-01-02"
	DefaultFormatZhTWShort    = "2006/1/2"
	DefaultFormatZhTWDateTime = "2006-01-02 15:04"
	DefaultFormatZhTWTime     = "15:04"

	DefaultFormatZhHKFull     = "年1月2日 Monday" // Chinese (Hong Kong)
	DefaultFormatZhHKLong     = "2006年1月2日"
	DefaultFormatZhHKMedium   = "2006-01-02"
	DefaultFormatZhHKShort    = "2006/1/2"
	DefaultFormatZhHKDateTime = "2006-01-02 15:04"
	DefaultFormatZhHKTime     = "15:04"

	DefaultFormatKoKRFull     = "년1월2일 월요일" // Korean (Korea)
	DefaultFormatKoKRLong     = "2006년1월2일"
	DefaultFormatKoKRMedium   = "2006-01-02"
	DefaultFormatKoKRShort    = "2006/1/2"
	DefaultFormatKoKRDateTime = "2006-01-02 15:04"
	DefaultFormatKoKRTime     = "15:04"

	DefaultFormatJaJPFull     = "年1月2日 Monday" // Japanese (Japan)
	DefaultFormatJaJPLong     = "2006年1月2日"
	DefaultFormatJaJPMedium   = "2006/01/02"
	DefaultFormatJaJPShort    = "2006/1/2"
	DefaultFormatJaJPDateTime = "2006/01/02 15:04"
	DefaultFormatJaJPTime     = "15:04"

	DefaultFormatElGRFull     = "Monday, 2 January" // Greek (Greece)
	DefaultFormatElGRLong     = "2 January 2006"
	DefaultFormatElGRMedium   = "2 Jan 2006"
	DefaultFormatElGRShort    = "02/01/06"
	DefaultFormatElGRDateTime = "02/01/06 15:04"
	DefaultFormatElGRTime     = "15:04"

	DefaultFormatCsCZFull     = "Monday, 2. January" // Czech (Czech Republic)
	DefaultFormatCsCZLong     = "2. January 2006"
	DefaultFormatCsCZMedium   = "02 Jan 2006"
	DefaultFormatCsCZShort    = "02/01/2006"
	DefaultFormatCsCZDateTime = "02/01/2006 15:04"
	DefaultFormatCsCZTime     = "15:04"

	DefaultFormatLtLTFull     = "m. January 2 d., Monday" // Lithuanian (Lithuania)
	DefaultFormatLtLTLong     = "2006 January 2 d."
	DefaultFormatLtLTMedium   = "2006 Jan 2"
	DefaultFormatLtLTShort    = "2006-01-02"
	DefaultFormatLtLTDateTime = "2006-01-02, 15:04"
	DefaultFormatLtLTTime     = "15:04"

	DefaultFormatThTHFull     = "Monday ที่ 2 January" // Thai (Thailand)
	DefaultFormatThTHLong     = "วันที่ 2 January 2006"
	DefaultFormatThTHMedium   = "2 Jan 2006"
	DefaultFormatThTHShort    = "02/01/2006"
	DefaultFormatThTHDateTime = "02/01/2006 15:04"
	DefaultFormatThTHTime     = "15:04"

	DefaultFormatUzUZFull     = "Monday, 02 January" // Uzbek (Uzbekistan)
	DefaultFormatUzUZLong     = "2 January 2006"
	DefaultFormatUzUZMedium   = "2 Jan 2006"
	DefaultFormatUzUZShort    = "02.01.2006"
	DefaultFormatUzUZDateTime = "02.01.2006 15:04"
	DefaultFormatUzUZTime     = "15:04"
)

// FullFormatsByLocale maps locales to the'full' date formats for all
// supported locales.
var fullFormatsByLocale = map[Locale]string{
	LocaleEn:   DefaultFormatEnUSFull,
	LocaleDe:   DefaultFormatDeDEFull,
	LocaleEs:   DefaultFormatEsESFull,
	LocaleFr:   DefaultFormatFrFRFull,
	LocaleIt:   DefaultFormatItITFull,
	LocaleHu:   DefaultFormatHuHUFull,
	LocaleNl:   DefaultFormatNlNLFull,
	LocalePl:   DefaultFormatPlFull,
	LocaleRo:   DefaultFormatRoROFull,
	LocaleSv:   DefaultFormatSvSEFull,
	LocaleTr:   DefaultFormatTrTRFull,
	LocaleBg:   DefaultFormatBgBGFull,
	LocaleRu:   DefaultFormatRuRUFull,
	LocaleFa:   DefaultFormatFaFull,
	LocaleKo:   DefaultFormatKoKRFull,
	LocaleJa:   DefaultFormatJaJPFull,
	LocaleUk:   DefaultFormatUkUAFull,
	LocaleEnAU: DefaultFormatEnUSFull,

	LocaleEnUS: DefaultFormatEnUSFull,
	LocaleEnGB: DefaultFormatEnGBFull,
	LocaleDaDK: DefaultFormatDaDKFull,
	LocaleNlBE: DefaultFormatNlBEFull,
	LocaleNlNL: DefaultFormatNlNLFull,
	LocaleFiFI: DefaultFormatFiFIFull,
	LocaleFrFR: DefaultFormatFrFRFull,
	LocaleFrCA: DefaultFormatFrCAFull,
	LocaleFrGP: DefaultFormatFrGPFull,
	LocaleFrLU: DefaultFormatFrLUFull,
	LocaleFrMQ: DefaultFormatFrMQFull,
	LocaleFrGF: DefaultFormatFrGFFull,
	LocaleFrRE: DefaultFormatFrREFull,
	LocaleDeDE: DefaultFormatDeDEFull,
	LocaleHuHU: DefaultFormatHuHUFull,
	LocaleItIT: DefaultFormatItITFull,
	LocaleNnNO: DefaultFormatNnNOFull,
	LocaleNbNO: DefaultFormatNbNOFull,
	LocalePtPT: DefaultFormatPtPTFull,
	LocalePtBR: DefaultFormatPtBRFull,
	LocaleRoRO: DefaultFormatRoROFull,
	LocaleRuRU: DefaultFormatRuRUFull,
	LocaleEsES: DefaultFormatEsESFull,
	LocaleCaES: DefaultFormatCaESFull,
	LocaleSvSE: DefaultFormatSvSEFull,
	LocaleTrTR: DefaultFormatTrTRFull,
	LocaleBgBG: DefaultFormatBgBGFull,
	LocaleZhCN: DefaultFormatZhCNFull,
	LocaleZhTW: DefaultFormatZhTWFull,
	LocaleZhHK: DefaultFormatZhHKFull,
	LocaleKoKR: DefaultFormatKoKRFull,
	LocaleJaJP: DefaultFormatJaJPFull,
	LocaleElGR: DefaultFormatElGRFull,
	LocaleCsCZ: DefaultFormatCsCZFull,
	LocaleUkUA: DefaultFormatUkUAFull,
	LocaleLtLT: DefaultFormatLtLTFull,
	LocaleThTH: DefaultFormatThTHFull,
	LocaleUzUZ: DefaultFormatUzUZFull,
}

// LongFormatsByLocale maps locales to the 'long' date formats for all
// supported locales.
var longFormatsByLocale = map[Locale]string{
	LocaleEn: DefaultFormatEnUSLong,
	LocaleDe: DefaultFormatDeDELong,
	LocaleEs: DefaultFormatEsESLong,
	LocaleFr: DefaultFormatFrFRLong,
	LocaleIt: DefaultFormatItITLong,
	LocaleNl: DefaultFormatNlNLLong,
	LocaleHu: DefaultFormatHuHULong,
	LocalePl: DefaultFormatPlLong,
	LocaleRo: DefaultFormatRoROLong,
	LocaleSv: DefaultFormatSvSELong,
	LocaleTr: DefaultFormatTrTRLong,
	LocaleBg: DefaultFormatBgBGLong,
	LocaleRu: DefaultFormatRuRULong,
	LocaleFa: DefaultFormatFaLong,
	LocaleJa: DefaultFormatJaJPLong,
	LocaleKo: DefaultFormatKoKRLong,
	LocaleUk: DefaultFormatUkUALong,

	LocaleEnAU: DefaultFormatEnUSLong,

	LocaleEnUS: DefaultFormatEnUSLong,
	LocaleEnGB: DefaultFormatEnGBLong,
	LocaleDaDK: DefaultFormatDaDKLong,
	LocaleNlBE: DefaultFormatNlBELong,
	LocaleNlNL: DefaultFormatNlNLLong,
	LocaleFiFI: DefaultFormatFiFILong,
	LocaleFrFR: DefaultFormatFrFRLong,
	LocaleFrCA: DefaultFormatFrCALong,
	LocaleFrGP: DefaultFormatFrGPLong,
	LocaleFrLU: DefaultFormatFrLULong,
	LocaleFrMQ: DefaultFormatFrMQLong,
	LocaleFrRE: DefaultFormatFrRELong,
	LocaleFrGF: DefaultFormatFrGFLong,
	LocaleDeDE: DefaultFormatDeDELong,
	LocaleHuHU: DefaultFormatHuHULong,
	LocaleItIT: DefaultFormatItITLong,
	LocaleNnNO: DefaultFormatNnNOLong,
	LocaleNbNO: DefaultFormatNbNOLong,
	LocalePtPT: DefaultFormatPtPTLong,
	LocalePtBR: DefaultFormatPtBRLong,
	LocaleRoRO: DefaultFormatRoROLong,
	LocaleRuRU: DefaultFormatRuRULong,
	LocaleEsES: DefaultFormatEsESLong,
	LocaleCaES: DefaultFormatCaESLong,
	LocaleSvSE: DefaultFormatSvSELong,
	LocaleTrTR: DefaultFormatTrTRLong,
	LocaleBgBG: DefaultFormatBgBGLong,
	LocaleZhCN: DefaultFormatZhCNLong,
	LocaleZhTW: DefaultFormatZhTWLong,
	LocaleZhHK: DefaultFormatZhHKLong,
	LocaleKoKR: DefaultFormatKoKRLong,
	LocaleJaJP: DefaultFormatJaJPLong,
	LocaleElGR: DefaultFormatElGRLong,
	LocaleCsCZ: DefaultFormatCsCZLong,
	LocaleUkUA: DefaultFormatUkUALong,
	LocaleLtLT: DefaultFormatLtLTLong,
	LocaleThTH: DefaultFormatThTHLong,
	LocaleUzUZ: DefaultFormatUzUZLong,
}

// MediumFormatsByLocale maps locales to the 'medium' date formats for all
// supported locales.
var mediumFormatsByLocale = map[Locale]string{
	LocaleEn: DefaultFormatEnUSMedium,
	LocaleDe: DefaultFormatDeDEMedium,
	LocaleEs: DefaultFormatEsESMedium,
	LocaleFr: DefaultFormatFrFRMedium,
	LocaleIt: DefaultFormatItITMedium,
	LocaleHu: DefaultFormatHuHUMedium,
	LocaleNl: DefaultFormatNlNLMedium,
	LocaleRo: DefaultFormatRoROMedium,
	LocalePl: DefaultFormatPlMedium,
	LocaleSv: DefaultFormatSvSEMedium,
	LocaleTr: DefaultFormatTrTRMedium,
	LocaleBg: DefaultFormatBgBGMedium,
	LocaleRu: DefaultFormatRuRUMedium,
	LocaleFa: DefaultFormatFaMedium,
	LocaleKo: DefaultFormatKoKRMedium,
	LocaleJa: DefaultFormatJaJPMedium,
	LocaleUk: DefaultFormatUkUAMedium,

	LocaleEnAU: DefaultFormatEnUSMedium,

	LocaleEnUS: DefaultFormatEnUSMedium,
	LocaleEnGB: DefaultFormatEnGBMedium,
	LocaleDaDK: DefaultFormatDaDKMedium,
	LocaleNlBE: DefaultFormatNlBEMedium,
	LocaleNlNL: DefaultFormatNlNLMedium,
	LocaleFiFI: DefaultFormatFiFIMedium,
	LocaleFrFR: DefaultFormatFrFRMedium,
	LocaleFrGP: DefaultFormatFrGPMedium,
	LocaleFrCA: DefaultFormatFrCAMedium,
	LocaleFrLU: DefaultFormatFrLUMedium,
	LocaleFrMQ: DefaultFormatFrMQMedium,
	LocaleFrGF: DefaultFormatFrGFMedium,
	LocaleFrRE: DefaultFormatFrREMedium,
	LocaleDeDE: DefaultFormatDeDEMedium,
	LocaleHuHU: DefaultFormatHuHUMedium,
	LocaleItIT: DefaultFormatItITMedium,
	LocaleNnNO: DefaultFormatNnNOMedium,
	LocaleNbNO: DefaultFormatNbNOMedium,
	LocalePtPT: DefaultFormatPtPTMedium,
	LocalePtBR: DefaultFormatPtBRMedium,
	LocaleRoRO: DefaultFormatRoROMedium,
	LocaleRuRU: DefaultFormatRuRUMedium,
	LocaleEsES: DefaultFormatEsESMedium,
	LocaleCaES: DefaultFormatCaESMedium,
	LocaleSvSE: DefaultFormatSvSEMedium,
	LocaleTrTR: DefaultFormatTrTRMedium,
	LocaleBgBG: DefaultFormatBgBGMedium,
	LocaleZhCN: DefaultFormatZhCNMedium,
	LocaleZhTW: DefaultFormatZhTWMedium,
	LocaleZhHK: DefaultFormatZhHKMedium,
	LocaleKoKR: DefaultFormatKoKRMedium,
	LocaleJaJP: DefaultFormatJaJPMedium,
	LocaleElGR: DefaultFormatElGRMedium,
	LocaleCsCZ: DefaultFormatCsCZMedium,
	LocaleUkUA: DefaultFormatUkUAMedium,
	LocaleLtLT: DefaultFormatLtLTMedium,
	LocaleThTH: DefaultFormatThTHMedium,
	LocaleUzUZ: DefaultFormatUzUZMedium,
}

// ShortFormatsByLocale maps locales to the 'short' date formats for all
// supported locales.
var shortFormatsByLocale = map[Locale]string{
	LocaleEn: DefaultFormatEnUSShort,
	LocaleDe: DefaultFormatDeDEShort,
	LocaleEs: DefaultFormatEsESShort,
	LocaleFr: DefaultFormatFrFRShort,
	LocaleIt: DefaultFormatItITShort,
	LocaleHu: DefaultFormatHuHUShort,
	LocaleNl: DefaultFormatNlNLShort,
	LocalePl: DefaultFormatPlShort,
	LocaleRo: DefaultFormatRoROShort,
	LocaleSv: DefaultFormatSvSEShort,
	LocaleTr: DefaultFormatTrTRShort,
	LocaleBg: DefaultFormatBgBGShort,
	LocaleRu: DefaultFormatRuRUShort,
	LocaleFa: DefaultFormatFaShort,
	LocaleKo: DefaultFormatKoKRShort,
	LocaleJa: DefaultFormatJaJPShort,
	LocaleUk: DefaultFormatUkUAShort,

	LocaleEnAU: DefaultFormatEnUSShort,

	LocaleEnUS: DefaultFormatEnUSShort,
	LocaleEnGB: DefaultFormatEnGBShort,
	LocaleDaDK: DefaultFormatDaDKShort,
	LocaleNlBE: DefaultFormatNlBEShort,
	LocaleNlNL: DefaultFormatNlNLShort,
	LocaleFiFI: DefaultFormatFiFIShort,
	LocaleFrFR: DefaultFormatFrFRShort,
	LocaleFrCA: DefaultFormatFrCAShort,
	LocaleFrLU: DefaultFormatFrLUShort,
	LocaleFrMQ: DefaultFormatFrMQShort,
	LocaleFrGF: DefaultFormatFrGFShort,
	LocaleFrGP: DefaultFormatFrGPShort,
	LocaleFrRE: DefaultFormatFrREShort,
	LocaleDeDE: DefaultFormatDeDEShort,
	LocaleHuHU: DefaultFormatHuHUShort,
	LocaleItIT: DefaultFormatItITShort,
	LocaleNnNO: DefaultFormatNnNOShort,
	LocaleNbNO: DefaultFormatNbNOShort,
	LocalePtPT: DefaultFormatPtPTShort,
	LocalePtBR: DefaultFormatPtBRShort,
	LocaleRoRO: DefaultFormatRoROShort,
	LocaleRuRU: DefaultFormatRuRUShort,
	LocaleEsES: DefaultFormatEsESShort,
	LocaleCaES: DefaultFormatCaESShort,
	LocaleSvSE: DefaultFormatSvSEShort,
	LocaleTrTR: DefaultFormatTrTRShort,
	LocaleBgBG: DefaultFormatBgBGShort,
	LocaleZhCN: DefaultFormatZhCNShort,
	LocaleZhTW: DefaultFormatZhTWShort,
	LocaleZhHK: DefaultFormatZhHKShort,
	LocaleKoKR: DefaultFormatKoKRShort,
	LocaleJaJP: DefaultFormatJaJPShort,
	LocaleElGR: DefaultFormatElGRShort,
	LocaleCsCZ: DefaultFormatCsCZShort,
	LocaleUkUA: DefaultFormatUkUAShort,
	LocaleLtLT: DefaultFormatLtLTShort,
	LocaleUzUZ: DefaultFormatUzUZShort,
}

// DateTimeFormatsByLocale maps locales to the 'DateTime' date formats for
// all supported locales.
var dateTimeFormatsByLocale = map[Locale]string{
	LocaleEn: DefaultFormatEnUSDateTime,
	LocaleDe: DefaultFormatDeDEDateTime,
	LocaleEs: DefaultFormatEsESDateTime,
	LocaleFr: DefaultFormatFrFRDateTime,
	LocaleIt: DefaultFormatItITDateTime,
	LocaleHu: DefaultFormatHuHUDateTime,
	LocaleNl: DefaultFormatNlNLDateTime,
	LocalePl: DefaultFormatPlDateTime,
	LocaleRo: DefaultFormatRoRODateTime,
	LocaleSv: DefaultFormatSvSEDateTime,
	LocaleTr: DefaultFormatTrTRDateTime,
	LocaleBg: DefaultFormatBgBGDateTime,
	LocaleRu: DefaultFormatRuRUDateTime,
	LocaleFa: DefaultFormatFaDateTime,
	LocaleKo: DefaultFormatKoKRDateTime,
	LocaleJa: DefaultFormatJaJPDateTime,
	LocaleUk: DefaultFormatUkUADateTime,

	LocaleEnAU: DefaultFormatEnUSDateTime,

	LocaleEnUS: DefaultFormatEnUSDateTime,
	LocaleEnGB: DefaultFormatEnGBDateTime,
	LocaleDaDK: DefaultFormatDaDKDateTime,
	LocaleNlBE: DefaultFormatNlBEDateTime,
	LocaleNlNL: DefaultFormatNlNLDateTime,
	LocaleFiFI: DefaultFormatFiFIDateTime,
	LocaleFrFR: DefaultFormatFrFRDateTime,
	LocaleFrCA: DefaultFormatFrCADateTime,
	LocaleFrGP: DefaultFormatFrGPDateTime,
	LocaleFrLU: DefaultFormatFrLUDateTime,
	LocaleFrMQ: DefaultFormatFrMQDateTime,
	LocaleFrGF: DefaultFormatFrGFDateTime,
	LocaleFrRE: DefaultFormatFrREDateTime,
	LocaleDeDE: DefaultFormatDeDEDateTime,
	LocaleHuHU: DefaultFormatHuHUDateTime,
	LocaleItIT: DefaultFormatItITDateTime,
	LocaleNnNO: DefaultFormatNnNODateTime,
	LocaleNbNO: DefaultFormatNbNODateTime,
	LocalePtPT: DefaultFormatPtPTDateTime,
	LocalePtBR: DefaultFormatPtBRDateTime,
	LocaleRoRO: DefaultFormatRoRODateTime,
	LocaleRuRU: DefaultFormatRuRUDateTime,
	LocaleEsES: DefaultFormatEsESDateTime,
	LocaleCaES: DefaultFormatCaESDateTime,
	LocaleSvSE: DefaultFormatSvSEDateTime,
	LocaleTrTR: DefaultFormatTrTRDateTime,
	LocaleBgBG: DefaultFormatBgBGDateTime,
	LocaleZhCN: DefaultFormatZhCNDateTime,
	LocaleZhTW: DefaultFormatZhTWDateTime,
	LocaleZhHK: DefaultFormatZhHKDateTime,
	LocaleKoKR: DefaultFormatKoKRDateTime,
	LocaleJaJP: DefaultFormatJaJPDateTime,
	LocaleElGR: DefaultFormatElGRDateTime,
	LocaleCsCZ: DefaultFormatCsCZDateTime,
	LocaleUkUA: DefaultFormatUkUADateTime,
	LocaleLtLT: DefaultFormatLtLTDateTime,
	LocaleUzUZ: DefaultFormatUzUZDateTime,
}

// TimeFormatsByLocale maps locales to the 'Time' date formats for
// all supported locales.
var TimeFormatsByLocale = map[Locale]string{
	LocaleEn: DefaultFormatEnUSTime,
	LocaleDe: DefaultFormatDeDETime,
	LocaleEs: DefaultFormatEsESTime,
	LocaleFr: DefaultFormatFrFRTime,
	LocaleIt: DefaultFormatItITTime,
	LocaleHu: DefaultFormatHuHUTime,
	LocaleNl: DefaultFormatNlNLTime,
	LocalePl: DefaultFormatPlTime,
	LocaleRo: DefaultFormatRoROTime,
	LocaleSv: DefaultFormatSvSETime,
	LocaleTr: DefaultFormatTrTRTime,
	LocaleBg: DefaultFormatBgBGTime,
	LocaleRu: DefaultFormatRuRUTime,
	LocaleFa: DefaultFormatFaTime,
	LocaleKo: DefaultFormatKoKRTime,
	LocaleJa: DefaultFormatJaJPTime,
	LocaleUk: DefaultFormatUkUATime,

	LocaleEnAU: DefaultFormatEnUSTime,

	LocaleEnUS: DefaultFormatEnUSTime,
	LocaleEnGB: DefaultFormatEnGBTime,
	LocaleDaDK: DefaultFormatDaDKTime,
	LocaleNlBE: DefaultFormatNlBETime,
	LocaleNlNL: DefaultFormatNlNLTime,
	LocaleFiFI: DefaultFormatFiFITime,
	LocaleFrFR: DefaultFormatFrFRTime,
	LocaleFrCA: DefaultFormatFrCATime,
	LocaleFrGP: DefaultFormatFrGPTime,
	LocaleFrLU: DefaultFormatFrLUTime,
	LocaleFrMQ: DefaultFormatFrMQTime,
	LocaleFrGF: DefaultFormatFrGFTime,
	LocaleFrRE: DefaultFormatFrRETime,
	LocaleDeDE: DefaultFormatDeDETime,
	LocaleHuHU: DefaultFormatHuHUTime,
	LocaleItIT: DefaultFormatItITTime,
	LocaleNnNO: DefaultFormatNnNOTime,
	LocaleNbNO: DefaultFormatNbNOTime,
	LocalePtPT: DefaultFormatPtPTTime,
	LocalePtBR: DefaultFormatPtBRTime,
	LocaleRoRO: DefaultFormatRoROTime,
	LocaleRuRU: DefaultFormatRuRUTime,
	LocaleEsES: DefaultFormatEsESTime,
	LocaleCaES: DefaultFormatCaESTime,
	LocaleSvSE: DefaultFormatSvSETime,
	LocaleTrTR: DefaultFormatTrTRTime,
	LocaleBgBG: DefaultFormatBgBGTime,
	LocaleZhCN: DefaultFormatZhCNTime,
	LocaleZhTW: DefaultFormatZhTWTime,
	LocaleZhHK: DefaultFormatZhHKTime,
	LocaleKoKR: DefaultFormatKoKRTime,
	LocaleJaJP: DefaultFormatJaJPTime,
	LocaleElGR: DefaultFormatElGRTime,
	LocaleCsCZ: DefaultFormatCsCZTime,
	LocaleUkUA: DefaultFormatUkUATime,
	LocaleLtLT: DefaultFormatLtLTTime,
	LocaleUzUZ: DefaultFormatUzUZTime,
}
