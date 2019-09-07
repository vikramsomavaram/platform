/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package tags

import (
	"github.com/osteele/liquid/render"
	"path/filepath"
)

func SectionTag(rc render.Context) (string, error) {
	value, err := rc.EvaluateString(rc.TagArgs())
	if err != nil {
		return "", err
	}
	rel, ok := value.(string)
	if !ok {
		return "", rc.Errorf("include requires a string argument; got %v", value)
	}
	filename := filepath.Join(filepath.Dir(rc.SourceFile()), rel)
	s, err := rc.RenderFile(filename, map[string]interface{}{})
	if err != nil {
		return "", err
	}
	return s, err
}
