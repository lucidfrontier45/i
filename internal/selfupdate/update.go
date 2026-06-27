// Package selfupdate replaces the running i binary with the latest release
// published on GitHub, including a Windows-safe replace that works around the
// locked running executable.
package selfupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

// ErrUpToDate is returned by SelfUpdate when the running version is already the
// latest release. It is not a failure.
var ErrUpToDate = errors.New("already up to date")

const githubAPI = "https://api.github.com/repos/"

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// SelfUpdate downloads the latest release of repo (an "owner/name" GitHub
// repository) and replaces the running executable with it. If currentVersion is
// already the latest, it returns ("", ErrUpToDate) without touching the binary.
// On success it returns the version that was installed.
func SelfUpdate(ctx context.Context, repo, currentVersion string) (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate executable: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", fmt.Errorf("resolve executable: %w", err)
	}

	// Best-effort cleanup of a previous run's leftover on Windows.
	_ = os.Remove(exe + ".to_remove")

	rel, err := LatestRelease(ctx, repo)
	if err != nil {
		return "", fmt.Errorf("fetch latest release: %w", err)
	}

	latestVersion := strings.TrimPrefix(rel.TagName, "v")

	if currentVersion != "dev" {
		if c := compareVersions(currentVersion, rel.TagName); c >= 0 {
			return "", ErrUpToDate
		}
	}

	name := assetName(latestVersion)
	downloadURL, ok := FindAsset(rel.Assets, name)
	if !ok {
		return "", fmt.Errorf(
			"no prebuilt binary for %s/%s (wanted asset %q)",
			runtime.GOOS,
			runtime.GOARCH,
			name,
		)
	}

	fmt.Printf("downloading %s...\n", downloadURL)
	data, err := DownloadURL(ctx, downloadURL)
	if err != nil {
		return "", fmt.Errorf("download %s: %w", name, err)
	}

	if err := verifyChecksum(ctx, repo, rel.TagName, name, data); err != nil {
		return "", fmt.Errorf("checksum %s: %w", name, err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(exe), ".i-update-*")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return "", fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("close temp file: %w", err)
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(tmpPath, 0o755); err != nil {
			return "", fmt.Errorf("chmod temp file: %w", err)
		}
	}

	if err := replaceExecutable(exe, tmpPath); err != nil {
		return "", fmt.Errorf("replace executable: %w", err)
	}

	return latestVersion, nil
}

// LatestRelease returns the latest non-prerelease of repo from the GitHub API.
func LatestRelease(ctx context.Context, repo string) (*Release, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		githubAPI+repo+"/releases/latest",
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"github api returned %s",
			resp.Status,
		)
	}

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

// ReleaseByTag returns the release with the given tag from the GitHub API.
func ReleaseByTag(ctx context.Context, repo, tag string) (*Release, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		githubAPI+repo+"/releases/tags/"+tag,
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"github api returned %s",
			resp.Status,
		)
	}

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

// assetName builds the goreleaser artifact name for the current platform, e.g.
// "i_0.1.0_linux_amd64" or "i_0.1.0_windows_amd64.exe".
func assetName(version string) string {
	name := fmt.Sprintf("i_%s_%s_%s", version, runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return name
}

// FindAsset returns the download URL of the asset whose name matches, if any.
func FindAsset(assets []Asset, name string) (string, bool) {
	for i := range assets {
		if assets[i].Name == name {
			return assets[i].BrowserDownloadURL, true
		}
	}
	return "", false
}

// DownloadURL fetches the full body of url.
func DownloadURL(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download returned %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// verifyChecksum downloads checksums.txt for tag and confirms the SHA256 of
// data matches the entry for assetName.
func verifyChecksum(
	ctx context.Context,
	repo,
	tag,
	assetName string,
	data []byte,
) error {
	url := fmt.Sprintf(
		"https://github.com/%s/releases/download/%s/checksums.txt",
		repo,
		tag,
	)
	body, err := DownloadURL(ctx, url)
	if err != nil {
		return fmt.Errorf("fetch checksums.txt: %w", err)
	}

	want, ok := parseChecksum(string(body), assetName)
	if !ok {
		return fmt.Errorf(
			"no checksum entry for %q in checksums.txt",
			assetName,
		)
	}

	sum := sha256.Sum256(data)
	got := hex.EncodeToString(sum[:])
	if !strings.EqualFold(got, want) {
		return fmt.Errorf(
			"checksum mismatch for %s: want %s, got %s",
			assetName,
			want,
			got,
		)
	}
	return nil
}

// parseChecksum scans a sha256sum-style checksums file (lines of
// "<hash>  <filename>") and returns the hash for name.
func parseChecksum(text, name string) (string, bool) {
	for _, line := range strings.Split(text, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		hash := fields[0]
		file := fields[len(fields)-1]
		if file == name {
			return hash, true
		}
	}
	return "", false
}

// compareVersions compares two versions using semver. Missing "v" prefixes are
// added; the result is -1, 0, or 1 (a < b, a == b, a > b).
func compareVersions(a, b string) int {
	return semver.Compare(normalizeSemver(a), normalizeSemver(b))
}

// normalizeSemver ensures a leading "v", as required by golang.org/x/mod/semver.
func normalizeSemver(v string) string {
	if !strings.HasPrefix(v, "v") {
		return "v" + v
	}
	return v
}

// replaceExecutable atomically installs newFile at exePath. On Windows the
// running executable is renamed out of the way first, since it cannot be
// overwritten or deleted while running; the leftover ".to_remove" file is removed on
// the next invocation.
func replaceExecutable(exePath, newFile string) error {
	if runtime.GOOS != "windows" {
		return os.Rename(newFile, exePath)
	}

	oldPath := exePath + ".to_remove"
	_ = os.Remove(oldPath) // clear a stale leftover from a much older run

	if err := os.Rename(exePath, oldPath); err != nil {
		return fmt.Errorf("rename running binary: %w", err)
	}

	if err := os.Rename(newFile, exePath); err != nil {
		// Put the old binary back so the tool stays usable.
		_ = os.Rename(oldPath, exePath)
		return fmt.Errorf("install new binary: %w", err)
	}

	// Still locked while this process runs; cleaned up next run.
	_ = os.Remove(oldPath)
	return nil
}
