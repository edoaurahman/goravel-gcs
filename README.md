# GCS Driver for Goravel

A Google Cloud Storage (GCS) disk driver for `facades.Storage()` of Goravel.

## Version Compatibility

| goravel/gcs | goravel/framework |
|-------------|-------------------|
| v1.0.*      | v1.14.*+          |

## Installation

Run the command below in your project to install the package automatically:

```bash
./artisan package:install github.com/edoaurahman/goravel-gcs
```

Or install manually:

```bash
go get github.com/edoaurahman/goravel-gcs
```

If you install manually, you need to:

1. Register the service provider in `config/app.go`:

```go
import "github.com/edoaurahman/goravel-gcs"

"providers": []foundation.ServiceProvider{
    // ...
    &goravelgcs.ServiceProvider{},
}
```

2. Publish the configuration file:

```bash
go run . artisan vendor:publish --package=github.com/edoaurahman/goravel-gcs
```

## Configuration

### 1. Setup GCS Credentials

First, create a service account in Google Cloud Console and download the JSON credentials file.

### 2. Configure Environment Variables

Add these variables to your `.env` file:

```env
GCS_PROJECT_ID=your-gcp-project-id
GCS_BUCKET=your-bucket-name
GCS_CREDENTIALS_PATH=/path/to/service-account-key.json
GCS_URL=https://storage.googleapis.com/your-bucket-name
```

### 3. Configure Filesystem Disk

Add the GCS disk configuration to `config/filesystems.go`:

```go
"gcs": map[string]any{
    "driver":      "gcs",
    "project_id":  config.Env("GCS_PROJECT_ID"),
    "bucket":      config.Env("GCS_BUCKET"),
    "credentials": config.Env("GCS_CREDENTIALS_PATH"),
    "url":         config.Env("GCS_URL"),
},
```

## Usage

```go
// Store a file
facades.Storage().Disk("gcs").Put("example.txt", "Contents")

// Get file contents
content, err := facades.Storage().Disk("gcs").Get("example.txt")

// Check if file exists
exists := facades.Storage().Disk("gcs").Exists("example.txt")

// Delete a file
err := facades.Storage().Disk("gcs").Delete("example.txt")

// Upload from HTTP request
file, err := ctx.Request().File("avatar")
path, err := file.Disk("gcs").Store("avatars")

// Get public URL
url := facades.Storage().Disk("gcs").Url("example.txt")

// Get temporary URL (signed URL)
url, err := facades.Storage().Disk("gcs").TemporaryUrl(
    "example.txt", 
    time.Now().Add(15*time.Minute),
)
```

## Testing

To run tests, set the following environment variables and run:

```bash
GCS_PROJECT_ID=your-project-id \
GCS_BUCKET=your-bucket \
GCS_CREDENTIALS_PATH=/path/to/credentials.json \
GCS_URL=https://storage.googleapis.com/your-bucket \
go test ./...
```

## Requirements

- Go 1.20+
- Goravel Framework v1.14+
- Google Cloud Storage account and bucket

## License

The GCS Driver for Goravel is open-sourced software licensed under the [MIT license](LICENSE).