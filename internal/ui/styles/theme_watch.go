package styles

// Hot reload of custom themes.
//
// Custom themes are read from the themes directory once at startup. A
// program that rewrites a theme file while slk is running (for example a
// desktop theme hook that exports the system palette as a slk theme)
// would otherwise only take effect after a restart. The UI polls
// ReloadCustomThemesIfChanged; when the directory changed and the active
// theme is a custom one, it calls Reapply.

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/gammons/slk/internal/core"
)

var (
	// The theme and overrides passed to the last Apply, so a reload can
	// re-apply the same selection with the new colors.
	activeName      string
	activeOverrides core.Theme

	// Fingerprint of the themes directory when it was last loaded.
	customFingerprint string
)

// customThemesFingerprint summarises the top-level .toml files in fsys by
// name, size and modification time. It changes whenever a theme file is
// added, removed or rewritten.
func customThemesFingerprint(fsys fs.FS) string {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return ""
	}
	parts := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".toml") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s:%d:%d", e.Name(), info.Size(), info.ModTime().UnixNano()))
	}
	sort.Strings(parts)
	return strings.Join(parts, "|")
}

// ReloadCustomThemesIfChanged re-reads the custom themes when the directory
// changed since it was last loaded, and reports whether it did.
func ReloadCustomThemesIfChanged(fsys fs.FS) bool {
	if customThemesFingerprint(fsys) == customFingerprint {
		return false
	}
	// Start from an empty registry so a deleted or renamed theme file
	// doesn't linger in the switcher.
	for k := range customThemes {
		delete(customThemes, k)
	}
	LoadCustomThemes(fsys)
	return true
}

// ActiveIsCustom reports whether the theme passed to the last Apply is a
// custom theme (as opposed to a built-in one).
func ActiveIsCustom() bool {
	_, ok := customThemes[strings.ToLower(activeName)]
	return ok
}

// Reapply applies the last selected theme again, picking up any change to
// its colors.
func Reapply() {
	Apply(activeName, activeOverrides)
}
