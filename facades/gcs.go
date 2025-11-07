package facades

import (
	"github.com/goravel/framework/contracts/filesystem"

	goravelgcs "github.com/edoaurahman/goravel-gcs"
)

func GCS(disk string) (filesystem.Driver, error) {
	instance, err := goravelgcs.App.MakeWith(goravelgcs.Binding, map[string]any{"disk": disk})
	if err != nil {
		return nil, err
	}

	return instance.(*goravelgcs.GCS), nil
}
