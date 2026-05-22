package goravelgcs

import (
	"context"
	"errors"
	"strings"
	"testing"

	"cloud.google.com/go/storage"
)

func TestCopyToDiskUsesServerSideCopier(t *testing.T) {
	originalResolveDiskConfig := defaultResolveDiskConfig
	originalRunCopierFrom := defaultRunCopierFrom
	defer func() {
		defaultResolveDiskConfig = originalResolveDiskConfig
		defaultRunCopierFrom = originalRunCopierFrom
	}()

	defaultResolveDiskConfig = func(disk string) (bucket, credentials string, err error) {
		if disk != "gcs_b" {
			t.Fatalf("unexpected destination disk: %s", disk)
		}

		return "bucket-b", "", nil
	}

	copierCalled := false
	defaultRunCopierFrom = func(_ context.Context, _ *storage.Client, srcBucket, srcFile, dstBucket, dstFile string) error {
		copierCalled = true

		if srcBucket != "bucket-a" {
			t.Fatalf("unexpected source bucket: %s", srcBucket)
		}
		if srcFile != "from/file.txt" {
			t.Fatalf("unexpected source file: %s", srcFile)
		}
		if dstBucket != "bucket-b" {
			t.Fatalf("unexpected destination bucket: %s", dstBucket)
		}
		if dstFile != "to/file.txt" {
			t.Fatalf("unexpected destination file: %s", dstFile)
		}

		return nil
	}

	gcs := &GCS{
		client: new(storage.Client),
		ctx:    context.Background(),
		disk:   "gcs_a",
		bucket: "bucket-a",
	}

	if err := gcs.CopyToDisk("gcs_b", "/from/file.txt", "/to/file.txt"); err != nil {
		t.Fatalf("copy to disk failed: %v", err)
	}

	if !copierCalled {
		t.Fatal("expected server-side copier to be called")
	}
}

func TestCopyToDiskRejectsDifferentCredentialFiles(t *testing.T) {
	originalResolveDiskConfig := defaultResolveDiskConfig
	originalRunCopierFrom := defaultRunCopierFrom
	defer func() {
		defaultResolveDiskConfig = originalResolveDiskConfig
		defaultRunCopierFrom = originalRunCopierFrom
	}()

	defaultResolveDiskConfig = func(disk string) (bucket, credentials string, err error) {
		return "bucket-b", "/tmp/destination-creds.json", nil
	}

	defaultRunCopierFrom = func(_ context.Context, _ *storage.Client, srcBucket, srcFile, dstBucket, dstFile string) error {
		return errors.New("copier should not be called")
	}

	gcs := &GCS{
		client:   new(storage.Client),
		ctx:      context.Background(),
		disk:     "gcs_a",
		bucket:   "bucket-a",
		credPath: "/tmp/source-creds.json",
	}

	err := gcs.CopyToDisk("gcs_b", "from/file.txt", "to/file.txt")
	if err == nil {
		t.Fatal("expected credential mismatch error")
	}

	if !strings.Contains(err.Error(), "different credentials files") {
		t.Fatalf("unexpected error: %v", err)
	}
}
