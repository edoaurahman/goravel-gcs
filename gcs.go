package goravelgcs

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/facades"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type GCS struct {
	client    *storage.Client
	bucket    string
	url       string
	ctx       context.Context
	disk      string
	projectID string
	credPath  string
}

func NewGCS() *GCS {
	return &GCS{
		ctx: context.Background(),
	}
}

func (g *GCS) WithContext(ctx context.Context) filesystem.Driver {
	g.ctx = ctx
	return g
}

// Initialize GCS client
func (g *GCS) init() error {
	if g.client != nil {
		return nil
	}

	// Read configuration
	g.projectID = facades.Config().GetString(fmt.Sprintf("filesystems.disks.%s.project_id", g.disk))
	g.bucket = facades.Config().GetString(fmt.Sprintf("filesystems.disks.%s.bucket", g.disk))
	g.credPath = facades.Config().GetString(fmt.Sprintf("filesystems.disks.%s.credentials", g.disk))
	g.url = facades.Config().GetString(fmt.Sprintf("filesystems.disks.%s.url", g.disk))

	// Default URL if not configured
	if g.url == "" {
		g.url = fmt.Sprintf("https://storage.googleapis.com/%s", g.bucket)
	}

	// Create GCS client
	var opts []option.ClientOption
	if g.credPath != "" {
		opts = append(opts, option.WithCredentialsFile(g.credPath))
	}

	client, err := storage.NewClient(g.ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to create GCS client: %w", err)
	}

	g.client = client
	return nil
}

func (g *GCS) AllDirectories(path string) ([]string, error) {
	if err := g.init(); err != nil {
		return nil, err
	}

	var directories []string
	seen := make(map[string]bool)

	it := g.client.Bucket(g.bucket).Objects(g.ctx, &storage.Query{
		Prefix:    g.normalizePath(path),
		Delimiter: "/",
	})

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		if attrs.Prefix != "" {
			dir := strings.TrimSuffix(attrs.Prefix, "/")
			if !seen[dir] {
				directories = append(directories, dir)
				seen[dir] = true
			}
		}
	}

	return directories, nil
}

func (g *GCS) AllFiles(path string) ([]string, error) {
	if err := g.init(); err != nil {
		return nil, err
	}

	var files []string

	it := g.client.Bucket(g.bucket).Objects(g.ctx, &storage.Query{
		Prefix: g.normalizePath(path),
	})

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		if !strings.HasSuffix(attrs.Name, "/") {
			files = append(files, attrs.Name)
		}
	}

	return files, nil
}

func (g *GCS) Copy(oldFile, newFile string) error {
	if err := g.init(); err != nil {
		return err
	}

	srcObject := g.client.Bucket(g.bucket).Object(g.normalizePath(oldFile))
	dstObject := g.client.Bucket(g.bucket).Object(g.normalizePath(newFile))

	_, err := dstObject.CopierFrom(srcObject).Run(g.ctx)
	return err
}

func (g *GCS) Delete(files ...string) error {
	if err := g.init(); err != nil {
		return err
	}

	for _, file := range files {
		obj := g.client.Bucket(g.bucket).Object(g.normalizePath(file))
		if err := obj.Delete(g.ctx); err != nil && err != storage.ErrObjectNotExist {
			return err
		}
	}
	return nil
}

func (g *GCS) DeleteDirectory(directory string) error {
	if err := g.init(); err != nil {
		return err
	}

	files, err := g.AllFiles(directory)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := g.Delete(file); err != nil {
			return err
		}
	}

	return nil
}

func (g *GCS) Directories(path string) ([]string, error) {
	if err := g.init(); err != nil {
		return nil, err
	}

	var directories []string
	seen := make(map[string]bool)

	normalizedPath := g.normalizePath(path)
	if normalizedPath != "" && !strings.HasSuffix(normalizedPath, "/") {
		normalizedPath += "/"
	}

	it := g.client.Bucket(g.bucket).Objects(g.ctx, &storage.Query{
		Prefix:    normalizedPath,
		Delimiter: "/",
	})

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		if attrs.Prefix != "" && attrs.Prefix != normalizedPath {
			dir := strings.TrimSuffix(attrs.Prefix, "/")
			if !seen[dir] {
				directories = append(directories, dir)
				seen[dir] = true
			}
		}
	}

	return directories, nil
}

func (g *GCS) Exists(file string) bool {
	if err := g.init(); err != nil {
		return false
	}

	obj := g.client.Bucket(g.bucket).Object(g.normalizePath(file))
	_, err := obj.Attrs(g.ctx)
	return err == nil
}

func (g *GCS) Files(path string) ([]string, error) {
	if err := g.init(); err != nil {
		return nil, err
	}

	var files []string

	normalizedPath := g.normalizePath(path)
	if normalizedPath != "" && !strings.HasSuffix(normalizedPath, "/") {
		normalizedPath += "/"
	}

	it := g.client.Bucket(g.bucket).Objects(g.ctx, &storage.Query{
		Prefix:    normalizedPath,
		Delimiter: "/",
	})

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		if attrs.Prefix == "" && !strings.HasSuffix(attrs.Name, "/") {
			files = append(files, attrs.Name)
		}
	}

	return files, nil
}

func (g *GCS) Get(file string) (string, error) {
	data, err := g.GetBytes(file)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (g *GCS) GetBytes(file string) ([]byte, error) {
	if err := g.init(); err != nil {
		return nil, err
	}

	obj := g.client.Bucket(g.bucket).Object(g.normalizePath(file))
	reader, err := obj.NewReader(g.ctx)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

func (g *GCS) LastModified(file string) (time.Time, error) {
	if err := g.init(); err != nil {
		return time.Time{}, err
	}

	obj := g.client.Bucket(g.bucket).Object(g.normalizePath(file))
	attrs, err := obj.Attrs(g.ctx)
	if err != nil {
		return time.Time{}, err
	}
	return attrs.Updated, nil
}

func (g *GCS) MakeDirectory(directory string) error {
	if err := g.init(); err != nil {
		return err
	}

	normalizedPath := g.normalizePath(directory)
	if !strings.HasSuffix(normalizedPath, "/") {
		normalizedPath += "/"
	}

	obj := g.client.Bucket(g.bucket).Object(normalizedPath)
	writer := obj.NewWriter(g.ctx)
	writer.ContentType = "application/x-directory"

	return writer.Close()
}

func (g *GCS) MimeType(file string) (string, error) {
	if err := g.init(); err != nil {
		return "", err
	}

	obj := g.client.Bucket(g.bucket).Object(g.normalizePath(file))
	attrs, err := obj.Attrs(g.ctx)
	if err != nil {
		return "", err
	}
	return attrs.ContentType, nil
}

func (g *GCS) Missing(file string) bool {
	return !g.Exists(file)
}

func (g *GCS) Move(oldFile, newFile string) error {
	if err := g.Copy(oldFile, newFile); err != nil {
		return err
	}
	return g.Delete(oldFile)
}

func (g *GCS) Path(file string) string {
	return g.normalizePath(file)
}

func (g *GCS) Put(file, content string) error {
	if err := g.init(); err != nil {
		return err
	}

	obj := g.client.Bucket(g.bucket).Object(g.normalizePath(file))
	writer := obj.NewWriter(g.ctx)

	if _, err := writer.Write([]byte(content)); err != nil {
		writer.Close()
		return err
	}

	return writer.Close()
}

func (g *GCS) PutFile(path string, source filesystem.File) (string, error) {
	filename := source.HashName()
	fullPath := filepath.Join(g.normalizePath(path), filename)

	return fullPath, g.putFileContent(fullPath, source)
}

func (g *GCS) PutFileAs(path string, source filesystem.File, name string) (string, error) {
	ext, err := source.Extension()
	if err != nil {
		return "", err
	}

	if !strings.Contains(name, ".") {
		name = name + "." + ext
	}

	fullPath := filepath.Join(g.normalizePath(path), name)
	return fullPath, g.putFileContent(fullPath, source)
}

func (g *GCS) putFileContent(path string, source filesystem.File) error {
	if err := g.init(); err != nil {
		return err
	}

	filePath := source.File()
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	obj := g.client.Bucket(g.bucket).Object(path)
	writer := obj.NewWriter(g.ctx)

	if mimeType, err := source.MimeType(); err == nil {
		writer.ContentType = mimeType
	}

	if _, err := io.Copy(writer, file); err != nil {
		writer.Close()
		return err
	}

	return writer.Close()
}

func (g *GCS) Size(file string) (int64, error) {
	if err := g.init(); err != nil {
		return 0, err
	}

	obj := g.client.Bucket(g.bucket).Object(g.normalizePath(file))
	attrs, err := obj.Attrs(g.ctx)
	if err != nil {
		return 0, err
	}
	return attrs.Size, nil
}

func (g *GCS) TemporaryUrl(file string, expiry time.Time) (string, error) {
	if err := g.init(); err != nil {
		return "", err
	}

	// First, check if the file exists
	object := g.client.Bucket(g.bucket).Object(g.normalizePath(file))
	_, err := object.Attrs(g.ctx)
	if err != nil {
		return "", fmt.Errorf("file not found: %w", err)
	}

	// Try to create signed URL with service account credentials
	if g.credPath != "" {
		opts := &storage.SignedURLOptions{
			Scheme:  storage.SigningSchemeV4,
			Method:  "GET",
			Expires: expiry,
		}

		// Load service account credentials
		googleAccessID, privateKey, err := g.loadCredentials()
		if err == nil {
			opts.GoogleAccessID = googleAccessID
			opts.PrivateKey = privateKey

			url, err := storage.SignedURL(g.bucket, g.normalizePath(file), opts)
			if err == nil {
				return url, nil
			}
		}
	}

	// Fallback: return public URL with warning
	// In production, you might want to make the bucket/objects public
	// or configure proper IAM permissions for signed URLs
	return g.Url(file), fmt.Errorf("signed URL not available, returning public URL (ensure your GCS bucket is public or configure proper credentials)")
}

func (g *GCS) loadCredentials() (string, []byte, error) {
	if g.credPath == "" {
		return "", nil, fmt.Errorf("credentials path not configured")
	}

	// Read the service account key file
	credData, err := os.ReadFile(g.credPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	// Parse the JSON to get the service account info
	var creds struct {
		ClientEmail string `json:"client_email"`
		PrivateKey  string `json:"private_key"`
	}

	if err := json.Unmarshal(credData, &creds); err != nil {
		return "", nil, fmt.Errorf("failed to parse credentials JSON: %w", err)
	}

	// Parse the private key
	block, _ := pem.Decode([]byte(creds.PrivateKey))
	if block == nil {
		return "", nil, fmt.Errorf("failed to decode PEM block from private key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return "", nil, fmt.Errorf("private key is not an RSA key")
	}

	// Convert RSA key to bytes
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(rsaKey)

	return creds.ClientEmail, privateKeyBytes, nil
}

func (g *GCS) generateSignedURLWithClient(file string, expiry time.Time) (string, error) {
	// Fallback approach: try to use Application Default Credentials
	// or return a public URL if signing is not possible

	// Check if the file exists first
	object := g.client.Bucket(g.bucket).Object(g.normalizePath(file))
	_, err := object.Attrs(g.ctx)
	if err != nil {
		return "", fmt.Errorf("file not found or not accessible: %w", err)
	}

	// For signed URLs without explicit credentials, we need to return
	// a public URL or implement IAM-based signing
	// For now, return public URL with a warning
	publicURL := g.Url(file)

	// You can uncomment this to see the fallback message
	// log.Printf("Warning: Using public URL instead of signed URL for %s", file)

	return publicURL, nil
}

func (g *GCS) Url(file string) string {
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(g.url, "/"), g.normalizePath(file))
}

func (g *GCS) normalizePath(path string) string {
	return strings.TrimPrefix(path, "/")
}
