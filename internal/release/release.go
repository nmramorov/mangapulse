package release

import (
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

// SelectNew выбирает новые релизы из переданных. Производит дедупликацию, проверку языка в релизах, проверку наличия в уже имеющихся релизах. Сортирует новые релизы по PublishedAt, при равенстве по Source, при равенстве по Source - по ID
func SelectNew(
	known map[Key]struct{},
	incoming []Release,
	language string,
) ([]Release, error) {
	if language == "" {
		return nil, fmt.Errorf("language could not be empty")
	}

	if len(incoming) == 0 {
		return nil, nil
	}

	err := validateIncomingReleases(incoming)
	if err != nil {
		return nil, err
	}

	filteredReleases := filterIncomingReleases(incoming, language, known)

	return sortReleases(filteredReleases), nil
}

func validateIncomingReleases(releases []Release) error {
	for _, release := range releases {
		if !release.Valid() {
			return fmt.Errorf("invalid release: %s", release.Key)
		}
	}

	return nil
}

func filterIncomingReleases(releases []Release, language string, known map[Key]struct{}) (filtered []Release) {
	seen := make(map[Key]struct{}, len(releases))
	// [en, ru] ru
	for _, release := range releases {
		key := release.Key
		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}

		if release.Language != language {
			continue
		}

		if _, ok := known[key]; ok {
			continue
		}

		filtered = append(filtered, release)
	}

	return
}

func sortReleases(releases []Release) []Release {
	sort.SliceStable(releases, func(i, j int) bool {
		if releases[i].PublishedAt.Before(releases[j].PublishedAt) {
			return true
		} else if releases[i].PublishedAt.After(releases[j].PublishedAt) {
			return false
		}

		if releases[i].Key.Source < releases[j].Key.Source {
			return true
		} else if releases[i].Key.Source > releases[j].Key.Source {
			return false
		}

		return releases[i].Key.ID < releases[j].Key.ID
	})

	return releases
}
