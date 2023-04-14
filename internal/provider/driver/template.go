package driver

import (
	"bytes"

	"go.uber.org/zap"
)

func (d *Driver) Template(name string, data any) string {
	result := &bytes.Buffer{}
	if err := d.tpl.ExecuteTemplate(result, name, data); err != nil {
		d.logger.Fatal("unable to execute template", zap.String("template", name), zap.Error(err))
	}
	return result.String()
}
