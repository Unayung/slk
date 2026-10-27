package emoji

import (
	"reflect"
	"testing"
)

func names(es []EmojiEntry) []string {
	out := make([]string, len(es))
	for i, e := range es {
		out[i] = e.Name
	}
	return out
}

func entryName(e EmojiEntry) string { return e.Name }

func entries(ns ...string) []EmojiEntry {
	out := make([]EmojiEntry, len(ns))
	for i, n := range ns {
		out[i] = EmojiEntry{Name: n}
	}
	return out
}

func TestFrecencyRank(t *testing.T) {
	f := NewFrecency([]string{"tada", "+1", "+1::skin-tone-3", "tada"})
	for _, tc := range []struct {
		name     string
		customs  map[string]string
		wantRank int
		wantOK   bool
	}{
		{"tada", nil, 0, true},     // a repeat keeps its first rank
		{"+1", nil, 1, true},       // recorded name
		{"thumbsup", nil, 1, true}, // alias of a recorded name
		{"+1_tone2", nil, 2, true}, // kyokomi spelling (tone N is Slack N+1)
		{"rocket", nil, 0, false},  // never used
		// A custom emoji named like a standard alias is its own emoji.
		{"thumbsup", map[string]string{"thumbsup": "https://e/x.png"}, 0, false},
	} {
		rank, ok := f.Rank(tc.name, tc.customs)
		if rank != tc.wantRank || ok != tc.wantOK {
			t.Errorf("Rank(%q) = %d, %v; want %d, %v", tc.name, rank, ok, tc.wantRank, tc.wantOK)
		}
	}
	if (Frecency{}).Len() != 0 {
		t.Error("zero Frecency ranks something")
	}
	if _, ok := (Frecency{}).Rank("tada", nil); ok {
		t.Error("zero Frecency ranked tada")
	}
}

func TestMatchPutsUsedEmojiFirst(t *testing.T) {
	all := entries("fire", "fire_engine", "fireworks", "campfire", "smile", "tada")
	use := NewFrecency([]string{"fireworks", "campfire", "fire_engine", "tada"})

	for _, tc := range []struct {
		name string
		q    string
		opt  MatchOptions
		want []string
	}{
		{
			name: "no history keeps alphabetical prefix order",
			q:    "fire",
			opt:  MatchOptions{},
			want: []string{"fire", "fire_engine", "fireworks"},
		},
		{
			name: "used prefix matches first, by frecency",
			q:    "fire",
			opt:  MatchOptions{Frecency: use},
			want: []string{"fireworks", "fire_engine", "fire"},
		},
		{
			name: "used substring matches come after used prefix matches, before unused ones",
			q:    "fire",
			opt:  MatchOptions{Frecency: use, Substring: true},
			want: []string{"fireworks", "fire_engine", "campfire", "fire"},
		},
		{
			name: "limit cuts the tail",
			q:    "fire",
			opt:  MatchOptions{Frecency: use, Substring: true, Limit: 2},
			want: []string{"fireworks", "fire_engine"},
		},
		{
			name: "empty query lists the history, then everything else",
			q:    "",
			opt:  MatchOptions{Frecency: use},
			want: []string{"fireworks", "campfire", "fire_engine", "tada", "fire", "smile"},
		},
		{
			name: "folds case",
			q:    "FIRE_",
			opt:  MatchOptions{Frecency: use},
			want: []string{"fire_engine"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := names(Match(all, entryName, tc.q, tc.opt))
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Match(%q) = %v, want %v", tc.q, got, tc.want)
			}
		})
	}
}

func TestMatchShowsAUsedEmojiOnce(t *testing.T) {
	// "+1" and "thumbsup" are the same emoji; it was used as "+1".
	all := entries("+1", "rocket", "thumbsup")
	use := NewFrecency([]string{"+1"})
	got := names(Match(all, entryName, "", MatchOptions{Frecency: use}))
	if want := []string{"+1", "rocket"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Match = %v, want %v", got, want)
	}
	// Typing an alias finds the used emoji under that alias.
	got = names(Match(BuildEntries(nil), entryName, "thumbs", MatchOptions{Frecency: use}))
	if len(got) == 0 || got[0] != "thumbsup" {
		t.Errorf(`Match("thumbs") = %v, want thumbsup first`, got)
	}
}

func TestMatchOnlyOffersEmojiInEntries(t *testing.T) {
	// A custom emoji used in another workspace is not in this one's
	// entries, so it is not offered.
	use := NewFrecency([]string{"other_workspace_parrot", "tada"})
	got := names(Match(entries("rocket", "tada"), entryName, "", MatchOptions{Frecency: use}))
	if want := []string{"tada", "rocket"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Match = %v, want %v", got, want)
	}
}

func TestUsedInText(t *testing.T) {
	customs := map[string]string{"partyparrot": "https://e/p.gif"}
	for _, tc := range []struct {
		in   string
		want []string
	}{
		{"ship it :rocket: :tada: :rocket:", []string{"rocket", "tada"}},
		{"nice :thumbsup:", []string{"+1"}},
		{"nice :+1::skin-tone-2:", []string{"+1::skin-tone-2"}},
		{"nice :thumbsup_tone3:", []string{"+1::skin-tone-4"}},
		{"meet at 12:30:45 :partyparrot:", []string{"partyparrot"}},
		{":not_an_emoji_name: plain", nil},
		{"", nil},
	} {
		if got := UsedInText(tc.in, customs); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("UsedInText(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
