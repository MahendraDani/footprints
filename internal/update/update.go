package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MahendraDani/footprints/internal/workspace"
	"golang.org/x/mod/semver"
)

const (
	Repo          = "MahendraDani/footprints"
	InstallURL    = "https://raw.githubusercontent.com/" + Repo + "/main/install.sh"
	checkInterval = 24 * time.Hour
	waitForCheck  = 800 * time.Millisecond // max extra time a command waits for the check
)

// state is the update-check cache stored in ~/.footprints/version.json.
type state struct {
	Latest    string    `json:"latest"`
	CheckedAt time.Time `json:"checked_at"`
}

func load() state {
	var st state
	p, err := workspace.VersionPath()
	if err != nil {
		return st
	}
	if b, err := os.ReadFile(p); err == nil {
		_ = json.Unmarshal(b, &st)
	}
	return st
}

func save(st state) {
	if _, err := workspace.EnsureRoot(); err != nil {
		return
	}
	p, err := workspace.VersionPath()
	if err != nil {
		return
	}
	b, err := json.Marshal(st)
	if err != nil {
		return
	}
	// Write then rename, so two footprints processes running at once can't leave a half-written file.
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return
	}
	_ = os.Rename(tmp, p)
}

func normalize(v string) string {
	if !strings.HasPrefix(v, "v") {
		return "v" + v
	}
	return v
}

func fetchLatest(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://api.github.com/repos/"+Repo+"/releases/latest", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "footprints-cli")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github api: %s", resp.Status)
	}
	var rel struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", err
	}
	return normalize(rel.TagName), nil
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

func disabled(args []string) bool {
	if os.Getenv("FOOTPRINTS_NO_UPDATE_CHECK") != "" || os.Getenv("CI") != "" || !isTerminal(os.Stderr) {
		return true
	}
	for _, a := range args {
		switch a {
		case "update", "completion", "help", "-v", "--version", "-h", "--help":
			return true
		}
	}
	return false
}

// Start begins a background update check (at most once per 24h) and returns a function.
// Call that function after the command finishes: it returns the notice to print, or "".
func Start(current string, args []string) func() string {
	cur := normalize(current)
	if !semver.IsValid(cur) || disabled(args) { // "dev" builds aren't valid semver, so they skip
		return func() string { return "" }
	}

	cached := load()
	var fresh chan state
	if time.Since(cached.CheckedAt) >= checkInterval {
		fresh = make(chan state, 1)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			st := cached
			if latest, err := fetchLatest(ctx); err == nil {
				st.Latest = latest
			}
			st.CheckedAt = time.Now() // on failure this backs off until tomorrow
			save(st)
			fresh <- st
		}()
	}

	return func() string {
		st := cached
		if fresh != nil {
			select {
			case st = <-fresh:
			case <-time.After(waitForCheck):
			}
		}
		if semver.IsValid(st.Latest) && semver.Compare(st.Latest, cur) > 0 {
			return fmt.Sprintf("\nA new version of footprints is available: %s → %s\nRun `footprints update` to upgrade.\n",
				strings.TrimPrefix(cur, "v"), strings.TrimPrefix(st.Latest, "v"))
		}
		return ""
	}
}
