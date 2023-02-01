package locales

import (
	"embed"
	"encoding/json"
	ginI18n "github.com/gin-contrib/i18n"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
)

//go:embed lang/*
var i18nFS embed.FS

type i18nEmbedLoader struct {
	FS embed.FS
}

func (c *i18nEmbedLoader) LoadMessage(path string) ([]byte, error) {
	return c.FS.ReadFile(path)
}

type MessageSource struct {
	C      *gin.Context
	Locale string
}

func (ms MessageSource) GetMessage(key string) string {
	bundle := ginI18n.WithBundle(defaultBundleConfig)
	handle := ginI18n.WithGetLngHandle(ms.getLangHandler)
	fn := ginI18n.Localize(bundle, handle)
	fn(ms.C)
	return ginI18n.MustGetMessage(key)
}

var defaultBundleConfig = &ginI18n.BundleCfg{
	RootPath:         "lang",
	AcceptLanguage:   []language.Tag{language.AmericanEnglish},
	DefaultLanguage:  language.AmericanEnglish,
	UnmarshalFunc:    json.Unmarshal,
	FormatBundleFile: "json",
	Loader:           &i18nEmbedLoader{FS: i18nFS},
}

func (ms MessageSource) getLangHandler(c *gin.Context, defaultLng string) string {
	lang := ms.Locale
	if c == nil || c.Request == nil || lang == "" {
		return defaultLng
	}
	return lang
}
