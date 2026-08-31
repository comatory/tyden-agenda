package agenda

import (
	"fmt"
	"strings"
)

const defaultLocale = "en"

type locale struct {
	tag          string
	translations map[string]string
}

func newLocale(tag string) (locale, error) {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		tag = defaultLocale
	}

	switch tag {
	case "en":
		return locale{tag: tag, translations: englishTranslations()}, nil
	case "cs":
		return locale{tag: tag, translations: czechTranslations()}, nil
	default:
		return locale{}, fmt.Errorf("unsupported locale %q", tag)
	}
}

func (locale locale) tagValue() string {
	return locale.tag
}

func (locale locale) formatTime(minutes int) string {
	return fmt.Sprintf("%02d:%02d", minutes/60, minutes%60)
}

func (locale locale) translate(key string) string {
	if value := locale.translations[key]; value != "" {
		return value
	}
	return key
}

func englishTranslations() map[string]string {
	return map[string]string{
		"day.mon": "Mon",
		"day.tue": "Tue",
		"day.wed": "Wed",
		"day.thu": "Thu",
		"day.fri": "Fri",
		"day.sat": "Sat",
		"day.sun": "Sun",
	}
}

func czechTranslations() map[string]string {
	return map[string]string{
		"day.mon": "Po",
		"day.tue": "Út",
		"day.wed": "St",
		"day.thu": "Čt",
		"day.fri": "Pá",
		"day.sat": "So",
		"day.sun": "Ne",
	}
}
