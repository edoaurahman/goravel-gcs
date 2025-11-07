package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

var config = `map[string]any{
        "driver": "custom",
        "project_id": config.Env("GCS_PROJECT_ID"),
        "bucket": config.Env("GCS_BUCKET"),
        "credentials": config.Env("GCS_CREDENTIALS_PATH"),
        "url": config.Env("GCS_URL"),
        "via": func() (filesystem.Driver, error) {
            return gcsfacades.GCS("gcs") // The ` + "`gcs`" + ` value is the ` + "`disks`" + ` key
        },
    }`

func main() {
	packages.Setup(os.Args).
		Install(
			modify.GoFile(path.Config("app.go")).
				Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.Register("&goravelgcs.ServiceProvider{}")),
			modify.GoFile(path.Config("filesystems.go")).
				Find(match.Imports()).Modify(modify.AddImport("github.com/goravel/framework/contracts/filesystem"), modify.AddImport("github.com/edoaurahman/goravel-gcs/facades", "gcsfacades")).
				Find(match.Config("filesystems.disks")).Modify(modify.AddConfig("gcs", config)).
				Find(match.Config("filesystems")).Modify(modify.AddConfig("default", `"gcs"`)),
		).
		Uninstall(
			modify.GoFile(path.Config("app.go")).
				Find(match.Providers()).Modify(modify.Unregister("&goravelgcs.ServiceProvider{}")).
				Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
			modify.GoFile(path.Config("filesystems.go")).
				Find(match.Config("filesystems.disks")).Modify(modify.RemoveConfig("gcs")).
				Find(match.Imports()).Modify(modify.RemoveImport("github.com/goravel/framework/contracts/filesystem"), modify.RemoveImport("github.com/edoaurahman/goravel-gcs/facades", "gcsfacades")).
				Find(match.Config("filesystems")).Modify(modify.AddConfig("default", `"local"`)),
		).
		Execute()
}
