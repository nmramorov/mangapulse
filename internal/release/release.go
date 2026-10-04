package release

import (
	"errors"
	"fmt"
	"sort"
	"time"
)

// Key ключ для релиза.
type Key struct {
	Source string
	ID     string
}

// Release описывает релиз манги.
type Release struct {
	Key         Key
	MangaID     string
	Chapter     string
	Language    string
	PublishedAt time.Time
}

// Valid проверка ключа.
func (K Key) Valid() bool {
	return K.Source != "" && K.ID != ""
}

// String строковое представление.
func (K Key) String() string {
	return fmt.Sprintf("%s-%s", K.Source, K.ID)
}

// Valid проверка релиза.
func (R Release) Valid() bool {
	return R.Key.Valid() && R.MangaID != "" && R.Language != "" && !R.PublishedAt.IsZero()
}

// SelectNew выбирает новые релизы из переданных.
func SelectNew(
	known map[Key]struct{},
	incoming []Release,
	language string,
) ([]Release, error) {
	if language == "" {
		return incoming, fmt.Errorf("language could not be empty")
	}

	err := validateIncomingReleases(incoming)
	if err != nil {
		return nil, err
	}

	filteredReleases := filterIncomingReleases(incoming, language, known)

	return sortReleases(filteredReleases), nil
}

func validateIncomingReleases(releases []Release) error {
	if len(releases) == 0 {
		return errors.New("releases were not provided")
	}

	for _, release := range releases {
		if !release.Valid() {
			return fmt.Errorf("invalid release: %s", release.Key)
		}
	}

	return nil
}

func filterIncomingReleases(releases []Release, language string, known map[Key]struct{}) (filtered []Release) {
	keysMap := make(map[Key]struct{}, len(releases))
	for _, release := range releases {
		if release.Language != language {
			continue
		}

		key := release.Key
		if _, ok := keysMap[key]; ok {
			continue
		}

		if _, ok := known[key]; ok {
			continue
		}

		keysMap[key] = struct{}{}

		filtered = append(filtered, release)
	}

	return
}

func sortReleases(releases []Release) []Release {
	sort.SliceStable(releases, func(i, j int) bool {
		if releases[i].PublishedAt.Before(releases[j].PublishedAt) {
			return true
		}

		if releases[i].Key.Source < releases[j].Key.Source {
			return true
		}

		return releases[i].Key.ID < releases[j].Key.ID
	})

	return releases
}
