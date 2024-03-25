package lpac

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

const (
	GitHubAPI = "https://api.github.com/repos/estkme-group/lpac/releases"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []GitHubReleaseAsset
}

type GitHubReleaseAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

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

	url, err := getDownloadUrl(version)
	if err != nil {
		slog.Info("failed to get download url", "version", version)
		return err
	}

	slog.Info("downloading lpac", "version", version, "url", url)
	return download(url, dataDir, version)
}

func download(url, dataDir, version string) error {
	if err := setupDstDir(dataDir); err != nil {
		return err
	}

	path, err := downloadFile(url, dataDir)
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

func getDownloadUrl(version string) (string, error) {
	resp, err := http.Get(GitHubAPI)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", err
	}
	for _, release := range releases {
		if release.TagName == version {
			for _, asset := range release.Assets {
				if asset.Name == packageNames[runtime.GOOS+":"+runtime.GOARCH] {
					return asset.URL, nil
				}
			}
		}
	}
	return "", errors.New("no download url found")
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
