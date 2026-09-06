package settings

import (
	"fmt"
	"strings"
	"text/template"

	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/failure"
)

type Locale string

const (
	LocaleEnglish Locale = "en"
	LocaleRussian Locale = "ru"
)

func Localize(key string) string {
	return LocalizeWith(key, Current.Locale, "")
}

func LocalizeWith(key string, locale Locale, grammarCase string) string {
	trans, err := cache.GetTranslation(fmt.Sprintf("assets/translations/strings_%v.toml", string(locale)))
	if err != nil {
		failure.LogErrWithLocation("failed to retrieve strings in %v for key %v: %v", locale, key, err)
		return "ERROR"
	}
	localizedText, ok := (*trans)[key+grammarCase]
	if !ok && len(grammarCase) > 0 {
		// When the key is not found with this grammar case, use the key by itself.
		localizedText, ok = (*trans)[key]
	}
	if !ok {
		// Fall back to English
		trans, err = cache.GetTranslation(fmt.Sprintf("assets/translations/strings_%v.toml", string(LocaleEnglish)))
		if err != nil {
			failure.LogErrWithLocation("failed to retrieve English fallback for localization key %v: %v", key, err)
			return "ERROR"
		}
		localizedText = (*trans)[key]
		if localizedText == "" {
			// There's no English translation, so it should show the key verbatim instead.
			// This is needed for things like Keyboard key bindings to show up correctly.
			return key
		}
	}

	// Parse as a text template in order to substitute control names and such.
	templ, err := template.New(key).
		Funcs(template.FuncMap{
			"acc": func(input any) string {
				switch in := input.(type) {
				case Action:
					return LocalizeWith(in.LocalizationKey(), locale, "Accusative")
				case string:
					return LocalizeWith(in, locale, "Accusative")
				}
				return "ERROR"
			},
		}).
		Parse(localizedText)

	if err != nil {
		failure.LogErrWithLocation("failed to parse template for key %v in lang %v: %v", key, locale, err)
		return "ERROR"
	}

	var finalText strings.Builder
	err = templ.Execute(&finalText, Current)
	if err != nil {
		failure.LogErrWithLocation("failed to execute template for key %v in lang %v: %v", key, locale, err)
		return "ERROR"
	}

	return finalText.String()
}

func (locale Locale) String() string {
	return LocalizeWith("myLang", locale, "")
}
