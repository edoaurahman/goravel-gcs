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

func GCSDriver(disk string) (*goravelgcs.GCS, error) {
	instance, err := goravelgcs.App.MakeWith(goravelgcs.Binding, map[string]any{"disk": disk})
	if err != nil {
		return nil, err
	}

	return instance.(*goravelgcs.GCS), nil
}

func CopyToDisk(sourceDisk, destinationDisk, oldFile, newFile string) error {
	driver, err := GCSDriver(sourceDisk)
	if err != nil {
		return err
	}

	return driver.CopyToDisk(destinationDisk, oldFile, newFile)
}

func MoveToDisk(sourceDisk, destinationDisk, oldFile, newFile string) error {
	driver, err := GCSDriver(sourceDisk)
	if err != nil {
		return err
	}

	return driver.MoveToDisk(destinationDisk, oldFile, newFile)
}
