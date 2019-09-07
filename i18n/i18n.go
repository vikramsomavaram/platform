/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package i18n

import (
	localization "github.com/nicksnyder/go-i18n/v2/i18n"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"
)

// I18nBundle ...
var I18nBundle *localization.Bundle

func init() {
	bundle := localization.NewBundle(language.English)

	//Load all language bundles
	//TODO load up from database using AddMessages method instead of files
	_, err := bundle.LoadMessageFile("es.toml")
	if err != nil {
		log.Error()
	}
	I18nBundle = bundle

}
