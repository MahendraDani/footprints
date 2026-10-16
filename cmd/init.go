package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/MahendraDani/footprints/internal/chrome"
	"github.com/MahendraDani/footprints/internal/workspace"
)

var (
	initProfile string
	initChoose  bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize footprints workspace and cache Chrome history",
	Long: `Creates ~/.footprints and copies your Chrome history file into it.

Chrome locks its history database while it is running, so footprints works on a
copy instead. Re-run this command to refresh the copy with your latest history.

If you use more than one Chrome profile, footprints asks which one to read and
remembers your choice. Use --choose to pick again, or --profile to skip the question.`,
	Example: `  footprints init
  footprints init --choose
  footprints init --profile "Profile 1"
  footprints init --profile work@example.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		profiles, err := chrome.List()
		if err != nil {
			return err
		}
		if len(profiles) == 0 {
			dir, _ := chrome.UserDataDir()
			return fmt.Errorf("no Chrome history found in %s\nIs Chrome installed, and have you browsed with it?", dir)
		}

		p, err := pickProfile(profiles)
		if err != nil {
			return err
		}

		if _, err := workspace.EnsureProfileDir(p.Dir); err != nil {
			return err
		}
		dst, err := workspace.ProfileHistoryPath(p.Dir)
		if err != nil {
			return err
		}
		if err := copyFile(p.Path, dst); err != nil {
			return fmt.Errorf("failed to copy history: %w", err)
		}

		// config.json is the dispatcher: without this, `graph` can't know which copy to read
		cfg := workspace.LoadConfig()
		cfg.ChromeProfile = p.Dir
		if err := workspace.SaveConfig(cfg); err != nil {
			return fmt.Errorf("copied history, but couldn't save your profile choice: %w", err)
		}

		fmt.Printf("Cached history from %s\n  → %s\n", describe(p), dst)
		if p.Visits == 0 {
			fmt.Println("Warning: this profile has no visits. Try `footprints init --choose`.")
		}
		for _, o := range profiles { // another profile with far more history is usually the one they meant
			if o.Dir != p.Dir && o.Visits > 100 && o.Visits > p.Visits*5 {
				fmt.Printf("Note: %s has far more history. Switch with `footprints init --profile %q`.\n", describe(o), o.Dir)
				break
			}
		}
		if len(profiles) > 1 {
			fmt.Println("Switch profiles any time with:  footprints init --choose")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVar(&initProfile, "profile", "", "Chrome profile to read: folder name (\"Profile 1\"), display name, or email")
	initCmd.Flags().BoolVar(&initChoose, "choose", false, "Choose the Chrome profile again, ignoring the saved choice")
}

func pickProfile(ps []chrome.Profile) (chrome.Profile, error) {
	if initProfile != "" {
		if p, ok := chrome.Resolve(ps, initProfile); ok {
			return p, nil
		}
		return chrome.Profile{}, fmt.Errorf("no Chrome profile matches %q\n%s", initProfile, profileList(ps))
	}
	if !initChoose {
		if saved := workspace.LoadConfig().ChromeProfile; saved != "" {
			if p, ok := chrome.Resolve(ps, saved); ok {
				return p, nil
			}
			fmt.Printf("Your saved profile %q no longer exists, so choosing again.\n", saved)
		}
	}
	if len(ps) == 1 {
		return ps[0], nil
	}
	// ps is sorted by most recent activity, so index 0 is the suggestion
	if interactive() {
		return prompt(ps)
	}
	fmt.Printf("Several Chrome profiles found; using %s. Pick another with --profile.\n", ps[0].Dir)
	return ps[0], nil
}

func prompt(ps []chrome.Profile) (chrome.Profile, error) {
	fmt.Println("Chrome profiles found:")
	fmt.Println()
	for i, p := range ps {
		fmt.Printf("  %d) %s\n", i+1, describe(p))
	}
	fmt.Printf("\nWhich profile's history should footprints use? [1] ")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && line == "" {
		return ps[0], nil // EOF: take the suggestion
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return ps[0], nil
	}
	if n, err := strconv.Atoi(line); err == nil && n >= 1 && n <= len(ps) {
		return ps[n-1], nil
	}
	if p, ok := chrome.Resolve(ps, line); ok {
		return p, nil
	}
	return chrome.Profile{}, fmt.Errorf("%q isn't one of the choices", line)
}

func describe(p chrome.Profile) string {
	s := p.Dir
	if p.Name != "" && p.Name != p.Dir {
		s += fmt.Sprintf("  %q", p.Name)
	}
	if p.Email != "" {
		s += " <" + p.Email + ">"
	}
	if p.Visits == 0 {
		return s + " — no visits"
	}
	last := p.Newest
	if len(last) >= 10 {
		last = last[:10]
	}
	return fmt.Sprintf("%s — %d visits, last %s", s, p.Visits, last)
}

func profileList(ps []chrome.Profile) string {
	var b strings.Builder
	b.WriteString("available profiles:")
	for _, p := range ps {
		b.WriteString("\n  " + describe(p))
	}
	return b.String()
}

func interactive() bool {
	in, e1 := os.Stdin.Stat()
	out, e2 := os.Stdout.Stat()
	return e1 == nil && e2 == nil && in.Mode()&os.ModeCharDevice != 0 && out.Mode()&os.ModeCharDevice != 0
}

// copyFile copies via a temp file, so an interrupted copy never leaves a half-written History behind.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	tmp := dst + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(tmp)
		return err
	}
	if err := out.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, dst)
}
