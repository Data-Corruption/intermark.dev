package templates

import (
	"context"
	"fmt"
	"html/template"

	"intermark/go/paths"

	"github.com/Data-Corruption/rlog/logger"
)

func Dict(values ...interface{}) map[string]interface{} {
	if len(values)%2 != 0 {
		panic("dict expects even number of args")
	}
	m := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			panic("dict keys must be strings")
		}
		m[key] = values[i+1]
	}
	return m
}

func LoadTemplates(ctx context.Context) (*template.Template, error) {
	funcs := template.FuncMap{"dict": Dict}
	tmpl := template.New("").Funcs(funcs)
	out, err := tmpl.ParseGlob(paths.TMPL_DIR + "/*.html")
	if err != nil {
		return nil, fmt.Errorf("error loading templates: %w", err)
	}
	logger.Debugf(ctx, "Loaded: %d templates\n", len(out.Templates()))
	return out, nil
}
