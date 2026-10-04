package release

import (
	"reflect"
	"testing"
	"time"
)

// Additional contract tests for MangaPulse PR #1, commit 749409e.
// Copy this file to internal/release/ and run:
//     go test ./internal/release -run '^TestReview' -count=1 -v
// These tests use only the standard library. They intentionally fail on the
// reviewed implementation. They do not prescribe an exported error name;
// add errors.Is assertions after choosing your validation error API.

func reviewValidRelease(source, id, language string) Release {
	return Release{
		Key:         Key{Source: source, ID: id},
		MangaID:     "manga-1",
		Chapter:     "42.5",
		Language:    language,
		PublishedAt: time.Date(2026, time.October, 1, 12, 0, 0, 0, time.UTC),
	}
}

func TestReviewEmptyInput(t *testing.T) {
	for _, tc := range []struct {
		name     string
		incoming []Release
	}{
		{name: "nil"},
		{name: "empty_non_nil", incoming: []Release{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SelectNew(nil, tc.incoming, "ru")
			if err != nil {
				t.Fatalf("empty input: unexpected error: %v", err)
			}
			if len(got) != 0 {
				t.Fatalf("empty input: got %d releases, want 0", len(got))
			}
		})
	}
}

func TestReviewEmptyLanguageHasNoResult(t *testing.T) {
	incoming := []Release{reviewValidRelease("source", "1", "ru")}
	got, err := SelectNew(nil, incoming, "")
	if err == nil {
		t.Fatal("empty language: expected an error")
	}
	if len(got) != 0 {
		t.Fatalf("error must not return releases: got %d", len(got))
	}
}

func TestReviewDeduplicateBeforeLanguage(t *testing.T) {
	first := reviewValidRelease("source", "1", "en")
	second := first
	second.Language = "ru"

	got, err := SelectNew(nil, []Release{first, second}, "ru")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("first occurrence is en: got %d releases, want 0", len(got))
	}
}

func TestReviewSortPriority(t *testing.T) {
	earlyBySource := reviewValidRelease("z", "1", "ru")
	lateBySource := reviewValidRelease("a", "2", "ru")
	lateBySource.PublishedAt = lateBySource.PublishedAt.Add(time.Hour)

	earlyByID := reviewValidRelease("source", "z", "ru")
	lateByID := reviewValidRelease("source", "a", "ru")
	lateByID.PublishedAt = lateByID.PublishedAt.Add(time.Hour)

	firstBySource := reviewValidRelease("a", "z", "ru")
	secondBySource := reviewValidRelease("z", "a", "ru")
	firstByID := reviewValidRelease("source", "a", "ru")
	secondByID := reviewValidRelease("source", "z", "ru")

	for _, tc := range []struct {
		name   string
		first  Release
		second Release
	}{
		{"date_before_source", earlyBySource, lateBySource},
		{"date_before_id", earlyByID, lateByID},
		{"source_before_id", firstBySource, secondBySource},
		{"id_tiebreaker", firstByID, secondByID},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, permutation := range []struct {
				name     string
				incoming []Release
			}{
				{"already_sorted", []Release{tc.first, tc.second}},
				{"reversed", []Release{tc.second, tc.first}},
			} {
				t.Run(permutation.name, func(t *testing.T) {
					got, err := SelectNew(nil, permutation.incoming, "ru")
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
					want := []Release{tc.first, tc.second}
					if !reflect.DeepEqual(got, want) {
						t.Fatalf("wrong sort order: got %v, want %v", reviewKeys(got), reviewKeys(want))
					}
				})
			}
		})
	}
}

func reviewKeys(releases []Release) []Key {
	keys := make([]Key, len(releases))
	for i, r := range releases {
		keys[i] = r.Key
	}
	return keys
}

func TestReviewValidateEachRequiredField(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*Release)
	}{
		{"source", func(r *Release) { r.Key.Source = "" }},
		{"id", func(r *Release) { r.Key.ID = "" }},
		{"manga_id", func(r *Release) { r.MangaID = "" }},
		{"language", func(r *Release) { r.Language = "" }},
		{"published_at", func(r *Release) { r.PublishedAt = time.Time{} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			valid := reviewValidRelease("source", "valid", "ru")
			invalid := reviewValidRelease("source", "invalid", "ru")
			tc.mutate(&invalid)
			got, err := SelectNew(nil, []Release{valid, invalid}, "ru")
			if err == nil {
				t.Fatal("expected validation error")
			}
			if len(got) != 0 {
				t.Fatalf("validation error: got %d releases, want no partial result", len(got))
			}
		})
	}
}

func TestReviewValidateBeforeDiscarding(t *testing.T) {
	valid := reviewValidRelease("source", "1", "ru")
	invalidOtherLanguage := reviewValidRelease("source", "2", "en")
	invalidOtherLanguage.MangaID = ""
	invalidDuplicate := valid
	invalidDuplicate.MangaID = ""
	invalidKnown := reviewValidRelease("source", "3", "ru")
	invalidKnown.MangaID = ""

	for _, tc := range []struct {
		name     string
		incoming []Release
		known    map[Key]struct{}
	}{
		{"wrong_language", []Release{valid, invalidOtherLanguage}, nil},
		{"duplicate", []Release{valid, invalidDuplicate}, nil},
		{"already_known", []Release{valid, invalidKnown}, map[Key]struct{}{invalidKnown.Key: {}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SelectNew(tc.known, tc.incoming, "ru")
			if err == nil || len(got) != 0 {
				t.Fatalf("must validate discarded records: got %d releases, err=%v", len(got), err)
			}
		})
	}
}

func TestReviewOwnership(t *testing.T) {
	first := reviewValidRelease("source", "a", "ru")
	second := reviewValidRelease("source", "b", "ru")
	incoming := []Release{second, first}
	before := append([]Release(nil), incoming...)
	knownKey := Key{Source: "source", ID: "known"}
	known := map[Key]struct{}{knownKey: {}}

	got, err := SelectNew(known, incoming, "ru")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(incoming, before) {
		t.Fatal("SelectNew changed incoming")
	}
	if !reflect.DeepEqual(known, map[Key]struct{}{knownKey: {}}) {
		t.Fatal("SelectNew changed known")
	}
	if len(got) != 2 {
		t.Fatalf("got %d releases, want 2", len(got))
	}
	got[0].Chapter = "changed"
	got[0].Key.Source = "changed"
	if !reflect.DeepEqual(incoming, before) {
		t.Fatal("mutating the result changed incoming")
	}
}

func TestReviewFirstDuplicateWins(t *testing.T) {
	first := reviewValidRelease("source", "1", "ru")
	second := first
	second.Chapter = "different"
	got, err := SelectNew(nil, []Release{first, second}, "ru")
	if err != nil || !reflect.DeepEqual(got, []Release{first}) {
		t.Fatalf("first duplicate must win: got=%+v, err=%v", got, err)
	}
}

func TestReviewAllKnown(t *testing.T) {
	r := reviewValidRelease("source", "1", "ru")
	got, err := SelectNew(map[Key]struct{}{r.Key: {}}, []Release{r}, "ru")
	if err != nil || len(got) != 0 {
		t.Fatalf("all known: got %d releases, err=%v", len(got), err)
	}
}

func TestReviewSameIDAcrossSources(t *testing.T) {
	first := reviewValidRelease("a", "1", "ru")
	second := reviewValidRelease("b", "1", "ru")
	got, err := SelectNew(nil, []Release{second, first}, "ru")
	if err != nil || !reflect.DeepEqual(got, []Release{first, second}) {
		t.Fatalf("ID is unique only within source: got=%+v, err=%v", got, err)
	}
}

func TestReviewEmptyChapterIsValid(t *testing.T) {
	r := reviewValidRelease("source", "1", "ru")
	r.Chapter = ""
	got, err := SelectNew(nil, []Release{r}, "ru")
	if err != nil || !reflect.DeepEqual(got, []Release{r}) {
		t.Fatalf("empty Chapter is permitted: got=%+v, err=%v", got, err)
	}
}
