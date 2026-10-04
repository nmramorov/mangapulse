package release

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const testTimeout = 1 * time.Second

func TestRelease_SelectNew(t *testing.T) {
	now := time.Now()

	type params struct {
		known    map[Key]struct{}
		incoming []Release
		language string
	}

	type test struct {
		name             string
		testParams       params
		errFunc          assert.ErrorAssertionFunc
		expectedReleases []Release
	}

	tests := []test{
		{
			name: "GIVEN language being an empty string THEN return error",
			testParams: params{
				language: "",
			},
			errFunc: assert.Error,
		},
		{
			name: "GIVEN 0 releases THEN return nil",
			testParams: params{
				language: "test",
			},
			errFunc: assert.NoError,
		},
		{
			name: "GIVEN invalid release ID THEN return error",
			testParams: params{
				language: "test",
				incoming: []Release{
					{
						Key: Key{
							ID:     "",
							Source: "test Source 1",
						},
						Language: "English",
					},
				},
			},
			errFunc: assert.Error,
		},
		{
			name: "GIVEN invalid release source THEN return error",
			testParams: params{
				language: "test",
				incoming: []Release{
					{
						Key: Key{
							ID:     "test ID",
							Source: "",
						},
						Language: "English",
					},
				},
			},
			errFunc: assert.Error,
		},
		{
			name: "GIVEN releases with various languages WHEN language set to 'English' THEN return only english releases",
			testParams: params{
				language: "English",
				incoming: []Release{
					{
						Key: Key{
							ID:     "test ID 1",
							Source: "test Source 1",
						},
						Language:    "English",
						MangaID:     "test",
						Chapter:     "test",
						PublishedAt: now,
					},
					{
						Key: Key{
							ID:     "test ID 2",
							Source: "test Source 2",
						},
						Language:    "Chinese",
						MangaID:     "test",
						Chapter:     "test",
						PublishedAt: now,
					},
				},
			},
			errFunc: assert.NoError,
			expectedReleases: []Release{
				{
					Key: Key{
						ID:     "test ID 1",
						Source: "test Source 1",
					},
					Language:    "English",
					MangaID:     "test",
					Chapter:     "test",
					PublishedAt: now,
				},
			},
		},
		{
			name: "GIVEN releases AND duplicates incoming THEN return only new without duplicates",
			testParams: params{
				language: "English",
				incoming: []Release{
					{
						Key: Key{
							ID:     "test ID 1",
							Source: "test Source 1",
						},
						Language:    "English",
						MangaID:     "test",
						Chapter:     "test",
						PublishedAt: now,
					},
					{
						Key: Key{
							ID:     "test ID 1",
							Source: "test Source 1",
						},
						Language:    "English",
						MangaID:     "test",
						Chapter:     "test",
						PublishedAt: now,
					},
					{
						Key: Key{
							ID:     "test ID 2",
							Source: "test Source 2",
						},
						Language:    "English",
						MangaID:     "test",
						Chapter:     "test",
						PublishedAt: now,
					},
				},
			},
			errFunc: assert.NoError,
			expectedReleases: []Release{
				{
					Key: Key{
						ID:     "test ID 1",
						Source: "test Source 1",
					},
					Language:    "English",
					MangaID:     "test",
					Chapter:     "test",
					PublishedAt: now,
				},
				{
					Key: Key{
						ID:     "test ID 2",
						Source: "test Source 2",
					},
					Language:    "English",
					MangaID:     "test",
					Chapter:     "test",
					PublishedAt: now,
				},
			},
		},
		{
			name: "GIVEN releases AND one already known THEN return only new",
			testParams: params{
				language: "English",
				known: map[Key]struct{}{
					{
						ID:     "test ID 1",
						Source: "test Source 1",
					}: {},
				},
				incoming: []Release{
					{
						Key: Key{
							ID:     "test ID 1",
							Source: "test Source 1",
						},
						Language:    "English",
						MangaID:     "test",
						Chapter:     "test",
						PublishedAt: now,
					},
					{
						Key: Key{
							ID:     "test ID 2",
							Source: "test Source 2",
						},
						Language:    "English",
						MangaID:     "test",
						Chapter:     "test",
						PublishedAt: now,
					},
				},
			},
			errFunc: assert.NoError,
			expectedReleases: []Release{
				{
					Key: Key{
						ID:     "test ID 2",
						Source: "test Source 2",
					},
					Language:    "English",
					MangaID:     "test",
					Chapter:     "test",
					PublishedAt: now,
				},
			},
		},
		{
			name: "GIVEN 2 releases with different PublishedAt THEN sorted properly",
			testParams: params{
				language: "English",
				incoming: []Release{
					{
						Key: Key{
							ID:     "test ID 1",
							Source: "test Source 1",
						},
						Language:    "English",
						MangaID:     "test",
						Chapter:     "test",
						PublishedAt: now,
					},
					{
						Key: Key{
							ID:     "test ID 2",
							Source: "test Source 2",
						},
						Language:    "English",
						MangaID:     "test",
						Chapter:     "test",
						PublishedAt: now.Add(-time.Second),
					},
				},
			},
			errFunc: assert.NoError,
			expectedReleases: []Release{
				{
					Key: Key{
						ID:     "test ID 2",
						Source: "test Source 2",
					},
					Language:    "English",
					MangaID:     "test",
					Chapter:     "test",
					PublishedAt: now.Add(-time.Second),
				},
				{
					Key: Key{
						ID:     "test ID 1",
						Source: "test Source 1",
					},
					Language:    "English",
					MangaID:     "test",
					Chapter:     "test",
					PublishedAt: now,
				},
			},
		},
		{
			name: "GIVEN 2 releases with different Sources THEN sorted properly",
			testParams: params{
				language: "English",
				incoming: []Release{
					{
						Key: Key{
							ID:     "test ID 1",
							Source: "test Source 1",
						},
						Language:    "English",
						MangaID:     "test",
						Chapter:     "test",
						PublishedAt: now,
					},
					{
						Key: Key{
							ID:     "test ID 2",
							Source: "a less test Source 2",
						},
						Language:    "English",
						MangaID:     "test",
						Chapter:     "test",
						PublishedAt: now,
					},
				},
			},
			errFunc: assert.NoError,
			expectedReleases: []Release{
				{
					Key: Key{
						ID:     "test ID 2",
						Source: "a less test Source 2",
					},
					Language:    "English",
					MangaID:     "test",
					Chapter:     "test",
					PublishedAt: now,
				},
				{
					Key: Key{
						ID:     "test ID 1",
						Source: "test Source 1",
					},
					Language:    "English",
					MangaID:     "test",
					Chapter:     "test",
					PublishedAt: now,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newReleases, err := SelectNew(tt.testParams.known, tt.testParams.incoming, tt.testParams.language)
			tt.errFunc(t, err)

			assert.EqualValues(t, tt.expectedReleases, newReleases)
		})
	}
}
