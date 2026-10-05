package styles

import (
	"testing"
	"testing/fstest"
	"time"

	"github.com/gammons/slk/internal/core"
)

func themeFile(name, primary string, mod time.Time) *fstest.MapFile {
	return &fstest.MapFile{
		Data:    []byte("name = \"" + name + "\"\n[colors]\nprimary = \"" + primary + "\"\n"),
		ModTime: mod,
	}
}

func TestReloadCustomThemesIfChanged(t *testing.T) {
	t0 := time.Unix(1_700_000_000, 0)
	fsys := fstest.MapFS{"omarchy.toml": themeFile("Omarchy", "#111111", t0)}

	LoadCustomThemes(fsys)
	Apply("Omarchy", core.Theme{})
	if !ActiveIsCustom() {
		t.Fatal("Omarchy should be a custom theme")
	}
	if got := lookupTheme("Omarchy").Primary; got != "#111111" {
		t.Fatalf("primary = %s, want #111111", got)
	}

	// Nothing changed: no reload.
	if ReloadCustomThemesIfChanged(fsys) {
		t.Fatal("reloaded without a change")
	}

	// The file was rewritten (new colors, new mtime): reload and re-apply.
	fsys["omarchy.toml"] = themeFile("Omarchy", "#222222", t0.Add(time.Second))
	if !ReloadCustomThemesIfChanged(fsys) {
		t.Fatal("did not notice the rewritten theme file")
	}
	Reapply()
	if got := lookupTheme("Omarchy").Primary; got != "#222222" {
		t.Fatalf("primary after reload = %s, want #222222", got)
	}
	// A deleted theme file disappears from the registry.
	delete(fsys, "omarchy.toml")
	fsys["other.toml"] = themeFile("Other", "#333333", t0)
	if !ReloadCustomThemesIfChanged(fsys) {
		t.Fatal("did not notice the removed theme file")
	}
	if ActiveIsCustom() {
		t.Fatal("Omarchy is gone, it should not count as custom any more")
	}
	for _, n := range ThemeNames() {
		if n == "Omarchy" {
			t.Fatal("deleted theme still listed")
		}
	}

	// A built-in theme is not affected by custom reloads.
	Apply("Dracula", core.Theme{})
	if ActiveIsCustom() {
		t.Fatal("Dracula is built in")
	}
}
