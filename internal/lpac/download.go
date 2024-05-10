package lpac

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

const (
	DownloadUrl = "https://github.com/estkme-group/lpac/releases/download/%s/%s"
)

var packageNames = map[string]string{
	"linux:amd64":   "lpac-linux-x86_64.zip",
	"windows:amd64": "lpac-windows-x86_64-mingw.zip",
	"windows:arm64": "lpac-windows-arm64-mingw.zip",
	"darwin:amd64":  "lpac-darwin-universal.zip",
	"darwin:arm64":  "lpac-darwin-universal.zip",
}

func Download(dataDir, version string) error {
	if _, ok := packageNames[runtime.GOOS+":"+runtime.GOARCH]; !ok {
		return errors.ErrUnsupported
	}

	if !shouldDownload(dataDir, version) {
		slog.Info("lpac already downloaded", "version", version)
		return nil
	}
	return download(dataDir, version)
}

func download(dataDir, version string) error {
	if err := setupDstDir(dataDir); err != nil {
		return err
	}

	path, err := downloadFile(fmt.Sprintf(DownloadUrl, version, packageNames[runtime.GOOS+":"+runtime.GOARCH]), dataDir)
	if err != nil {
		return err
	}

	if err := unzip(path, dataDir); err != nil {
		return err
	}

	if _, err := os.Create(filepath.Join(dataDir, version)); err != nil {
		slog.Warn("failed to create version file", "version", version, "error", err)
	}

	if err := os.Remove(path); err != nil {
		slog.Warn("failed to remove zip file", "path", path, "error", err)
	}
	return nil
}

func unzip(filePath string, dataDir string) error {
	zipReader, err := zip.OpenReader(filePath)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		if f.FileInfo().IsDir() {
			continue
		}

		dst, err := os.OpenFile(filepath.Join(dataDir, f.Name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		src, err := f.Open()
		if err != nil {
			return err
		}
		io.Copy(dst, src)
		src.Close()
		dst.Close()
	}
	return nil
}

func downloadFile(url string, dataDir string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	filePath := filepath.Join(dataDir, "lpac.zip")
	out, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return filePath, err
}

func setupDstDir(dataDir string) error {
	os.RemoveAll(dataDir)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}
	return nil
}

func shouldDownload(dataDir, version string) bool {
	versionFile := filepath.Join(dataDir, version)
	stat, err := os.Stat(versionFile)
	if os.IsNotExist(err) {
		return true
	}
	if stat.Name() != version {
		return true
	}
	return false
}
