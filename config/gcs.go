package config

import (
	"github.com/goravel/framework/facades"
)

func init() {
	config := facades.Config()
	config.Add("gcs", map[string]any{
		// GCS Project ID
		"project_id": config.Env("GCS_PROJECT_ID"),

		// GCS Bucket Name
		"bucket": config.Env("GCS_BUCKET"),

		// Path to service account credentials JSON file
		"credentials": config.Env("GCS_CREDENTIALS_PATH"),

		// Public URL for the bucket
		// Leave empty to use default: https://storage.googleapis.com/{bucket}
		"url": config.Env("GCS_URL"),
	})
}
