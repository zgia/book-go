// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package translation

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"golang.org/x/text/language"

	"zgia.net/book/internal/conf"
	log "zgia.net/book/internal/logger"
	"zgia.net/book/internal/modules/translation/i18n"
	"zgia.net/book/internal/util"
)

// Locale represents an interface to translation
type Locale interface {
	Language() string
	Tr(string, ...interface{}) string
	TrN(cnt interface{}, key1, keyN string, args ...interface{}) string
}

// LangType represents a lang type
type LangType struct {
	Lang, Name string // these fields are used directly in templates: {{range .AllLangs}}{{.Lang}}{{.Name}}{{end}}
}

var (
	lock          *sync.RWMutex
	matcher       language.Matcher
	allLangs      []*LangType
	allLangMap    map[string]*LangType
	supportedTags []language.Tag
)

var Lang Locale

// AllLangs returns all supported languages sorted by name
func AllLangs() []*LangType {
	return allLangs
}

// InitLocales loads the locales
func InitLocales() {
	if lock != nil {
		lock.Lock()
		defer lock.Unlock()
	} else if !conf.IsProdMode() && lock == nil {
		lock = &sync.RWMutex{}
	}

	refreshLocales := func() {
		i18n.ResetDefaultLocales()

		localeFiles, err := filepath.Glob(filepath.Join(util.CustomDir(), "conf", "locale", "*.ini"))
		if err != nil {
			log.Fatalf("Failed to list locale files: %v", err)
		}

		localeData := make(map[string][]byte, len(localeFiles))
		for _, name := range localeFiles {
			localeData[path.Base(name)], err = os.ReadFile(name)
			if err != nil {
				log.Fatalf("Failed to load %s locale file. %v", name, err)
			}
		}

		supportedTags = make([]language.Tag, len(conf.I18n.Langs))
		for i, lang := range conf.I18n.Langs {
			supportedTags[i] = language.Raw.Make(lang)
		}

		matcher = language.NewMatcher(supportedTags)
		for i := range conf.I18n.Names {
			var localeDataBase []byte
			if i == 0 && conf.I18n.Langs[0] != "en-US" {
				// Only en-US has complete translations. When use other language as default, the en-US should still be used as fallback.
				localeDataBase = localeData["locale_en-US.ini"]
				if localeDataBase == nil {
					log.Fatal("Failed to load locale_en-US.ini file.")
				}
			}

			key := "locale_" + conf.I18n.Langs[i] + ".ini"
			if err = i18n.DefaultLocales.AddLocaleByIni(conf.I18n.Langs[i], conf.I18n.Names[i], localeDataBase, localeData[key]); err != nil {
				log.Errorf("Failed to set messages to %s: %v", conf.I18n.Langs[i], err)
			}
		}
		if len(conf.I18n.Langs) != 0 {
			defaultLangName := conf.I18n.Langs[0]
			if defaultLangName != "en-US" {
				log.Infof("Use the first locale (%s) in LANGS setting option as default", defaultLangName)
			}
			i18n.DefaultLocales.SetDefaultLang(defaultLangName)
		}
	}

	refreshLocales()

	langs, descs := i18n.DefaultLocales.ListLangNameDesc()
	allLangs = make([]*LangType, 0, len(langs))
	allLangMap = map[string]*LangType{}
	for i, v := range langs {
		l := &LangType{v, descs[i]}
		allLangs = append(allLangs, l)
		allLangMap[v] = l
	}

	// Sort languages case-insensitive according to their name - needed for the user settings
	sort.Slice(allLangs, func(i, j int) bool {
		return strings.ToLower(allLangs[i].Name) < strings.ToLower(allLangs[j].Name)
	})
}

// Match matches accept languages
func Match(tags ...language.Tag) language.Tag {
	_, i, _ := matcher.Match(tags...)
	return supportedTags[i]
}

// locale represents the information of localization.
type locale struct {
	i18n.Locale
	Lang, LangName string // these fields are used directly in templates: .i18n.Lang
}

// NewLocale return a locale
func NewLocale(lang string) Locale {
	if lock != nil {
		lock.RLock()
		defer lock.RUnlock()
	}

	langName := "unknown"
	if l, ok := allLangMap[lang]; ok {
		langName = l.Name
	}
	i18nLocale, _ := i18n.GetLocale(lang)
	return &locale{
		Locale:   i18nLocale,
		Lang:     lang,
		LangName: langName,
	}
}

func (l *locale) Language() string {
	return l.Lang
}

// Language specific rules for translating plural texts
var trNLangRules = map[string]func(int64) int{
	// the default rule is "en-US" if a language isn't listed here
	"en-US": func(cnt int64) int {
		if cnt == 1 {
			return 0
		}
		return 1
	},
	"lv-LV": func(cnt int64) int {
		if cnt%10 == 1 && cnt%100 != 11 {
			return 0
		}
		return 1
	},
	"ru-RU": func(cnt int64) int {
		if cnt%10 == 1 && cnt%100 != 11 {
			return 0
		}
		return 1
	},
	"zh-CN": func(cnt int64) int {
		return 0
	},
	"zh-HK": func(cnt int64) int {
		return 0
	},
	"zh-TW": func(cnt int64) int {
		return 0
	},
	"fr-FR": func(cnt int64) int {
		if cnt > -2 && cnt < 2 {
			return 0
		}
		return 1
	},
}

// TrN returns translated message for plural text translation
func (l *locale) TrN(cnt interface{}, key1, keyN string, args ...interface{}) string {
	var c int64
	if t, ok := cnt.(int); ok {
		c = int64(t)
	} else if t, ok := cnt.(int16); ok {
		c = int64(t)
	} else if t, ok := cnt.(int32); ok {
		c = int64(t)
	} else if t, ok := cnt.(int64); ok {
		c = t
	} else {
		return l.Tr(keyN, args...)
	}

	ruleFunc, ok := trNLangRules[l.Lang]
	if !ok {
		ruleFunc = trNLangRules["en-US"]
	}

	if ruleFunc(c) == 0 {
		return l.Tr(key1, args...)
	}
	return l.Tr(keyN, args...)
}
