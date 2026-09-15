package system

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	// Mockable URL and HTTP client for testing
	ModelBundleURL = "https://github.com/KrushnaVardhanReddy/local-ai-assistant/releases/download/v1.0-models/sherpa-onnx-parakeet-bundle.zip"
	HTTPClient     = &http.Client{Timeout: 30 * time.Minute}
)

type progressWriter struct {
	total    float64
	written  float64
	callback func(float32)
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.written += float64(n)
	if pw.total > 0 && pw.callback != nil {
		pw.callback(float32(pw.written / pw.total))
	}
	return n, nil
}

// StartBackgroundDownload runs in a goroutine and downloads the Parakeet model bundle if supported.
func StartBackgroundDownload(ctx context.Context, progressCallback func(float32)) {
	if !IsParakeetSupported() {
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	modelsDir := filepath.Join(homeDir, ".local", "share", "barnowl", "models")
	parakeetDir := filepath.Join(modelsDir, "parakeet")

	// Check for a specific required file to be sure download completed fully
	targetFile := filepath.Join(parakeetDir, "model.onnx")
	if info, err := os.Stat(targetFile); err == nil && !info.IsDir() {
		return
	}

	// Create models directory if it doesn't exist
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		return
	}

	zipPath := filepath.Join(modelsDir, "sherpa-onnx-parakeet-bundle.zip")

	err = downloadFileWithRetry(ctx, ModelBundleURL, zipPath, progressCallback, 3)
	if err != nil {
		return
	}

	// Extract the zip
	err = extractZip(zipPath, modelsDir)
	if err != nil {
		return
	}

	// Clean up zip file
	_ = os.Remove(zipPath)
}

func downloadFileWithRetry(ctx context.Context, url string, dest string, progressCallback func(float32), maxRetries int) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		err := downloadFile(ctx, url, dest, progressCallback)
		if err == nil {
			return nil
		}
		lastErr = err

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second): // basic backoff
		}
	}
	return fmt.Errorf("failed to download after %d retries: %v", maxRetries, lastErr)
}

func DownloadFileAtomic(ctx context.Context, url string, dest string, progressCallback func(float32)) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	tmpDest := dest + ".tmp"
	out, err := os.Create(tmpDest)
	if err != nil {
		return err
	}

	pw := &progressWriter{
		total:    float64(resp.ContentLength),
		callback: progressCallback,
	}

	_, err = io.Copy(out, io.TeeReader(resp.Body, pw))
	out.Close()
	if err != nil {
		os.Remove(tmpDest)
		return err
	}

	if err := os.Rename(tmpDest, dest); err != nil {
		os.Remove(tmpDest)
		return err
	}

	return nil
}

func downloadFile(ctx context.Context, url string, dest string, progressCallback func(float32)) error {
	return DownloadFileAtomic(ctx, url, dest, progressCallback)
}

func extractZip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// Prevent ZipSlip vulnerability
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}
