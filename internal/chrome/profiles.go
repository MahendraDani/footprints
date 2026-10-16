package chrome

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

// Stats summarizes one History database.
type Stats struct {
	Visits, Linked int
	Oldest, Newest string // local time, "YYYY-MM-DD HH:MM:SS"
}

// Profile is one Chrome profile that has a History file.
type Profile struct {
	Dir   string // folder name, e.g. "Default" or "Profile 1"
	Name  string // display name from Chrome, e.g. "Work"
	Email string
	Path  string // full path to the History file
	Stats
}

// UserDataDir returns Chrome's user-data folder for this OS.
func UserDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Google", "Chrome"), nil
	case "windows":
		return filepath.Join(home, "AppData", "Local", "Google", "Chrome", "User Data"), nil
	case "linux":
		return filepath.Join(home, ".config", "google-chrome"), nil
	}
	return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
}

type infoEntry struct {
	Name     string `json:"name"`
	UserName string `json:"user_name"`
}

// readLocalState reads profile display names and emails from Chrome's "Local State" file.
func readLocalState(base string) map[string]infoEntry {
	var ls struct {
		Profile struct {
			InfoCache map[string]infoEntry `json:"info_cache"`
		} `json:"profile"`
	}
	if b, err := os.ReadFile(filepath.Join(base, "Local State")); err == nil {
		_ = json.Unmarshal(b, &ls)
	}
	return ls.Profile.InfoCache
}

// StatHistory inspects a History file. It works on a temp copy because Chrome locks the live file.
func StatHistory(path string) (Stats, error) {
	var s Stats
	src, err := os.Open(path)
	if err != nil {
		return s, err
	}
	defer src.Close()
	tmp, err := os.CreateTemp("", "footprints-stat-*.db")
	if err != nil {
		return s, err
	}
	defer os.Remove(tmp.Name())
	if _, err := io.Copy(tmp, src); err != nil {
		tmp.Close()
		return s, err
	}
	tmp.Close()

	db, err := sql.Open("sqlite", tmp.Name())
	if err != nil {
		return s, err
	}
	defer db.Close()

	var o, n sql.NullString
	if err := db.QueryRow(`SELECT count(*),
		datetime(min(visit_time)/1000000-11644473600,'unixepoch','localtime'),
		datetime(max(visit_time)/1000000-11644473600,'unixepoch','localtime')
		FROM visits`).Scan(&s.Visits, &o, &n); err != nil {
		return s, err
	}
	s.Oldest, s.Newest = o.String, n.String
	_ = db.QueryRow(`SELECT count(*) FROM visits v JOIN visits p ON p.id = v.from_visit`).Scan(&s.Linked)
	return s, nil
}

// List returns every profile with a History file, most recently active first.
func List() ([]Profile, error) {
	base, err := UserDataDir()
	if err != nil {
		return nil, err
	}
	matches, _ := filepath.Glob(filepath.Join(base, "*", "History"))
	info := readLocalState(base)

	var out []Profile
	for _, m := range matches {
		dir := filepath.Base(filepath.Dir(m))
		if dir == "Guest Profile" || dir == "System Profile" {
			continue
		}
		p := Profile{Dir: dir, Path: m}
		if i, ok := info[dir]; ok {
			p.Name, p.Email = i.Name, i.UserName
		}
		if s, err := StatHistory(m); err == nil {
			p.Stats = s
		}
		out = append(out, p)
	}
	sort.SliceStable(out, func(a, b int) bool { return out[a].Newest > out[b].Newest })
	return out, nil
}

// Resolve finds a profile by folder name, display name, or email (case-insensitive).
func Resolve(ps []Profile, q string) (Profile, bool) {
	q = strings.ToLower(strings.TrimSpace(q))
	for _, p := range ps {
		if strings.ToLower(p.Dir) == q {
			return p, true
		}
	}
	for _, p := range ps {
		if strings.ToLower(p.Name) == q || (p.Email != "" && strings.ToLower(p.Email) == q) {
			return p, true
		}
	}
	return Profile{}, false
}
