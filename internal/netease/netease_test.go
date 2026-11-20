package netease

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestPlayableFollowsPrivilege(t *testing.T) {
	tracks := []Track{{ID: 1, Fee: 1}, {ID: 2, Fee: 8}, {ID: 3, Fee: 8}, {ID: 4, Fee: 1}}
	if tracks[0].Playable() || !tracks[1].Playable() {
		t.Fatal("without privileges, VIP songs do not play and free ones do")
	}
	withPrivileges(tracks, []Privilege{
		{ID: 1, Pl: 320000},           // VIP song, the listener subscribes
		{ID: 2, Pl: 0},                // nothing to play
		{ID: 3, St: -200, Pl: 128000}, // taken down
	})
	want := []bool{true, false, false, false}
	for i, tr := range tracks {
		if tr.Playable() != want[i] {
			t.Errorf("track %d: playable %v", tr.ID, tr.Playable())
		}
	}
}

// 风控和登录错误曾被当成 VIP 限制；保留业务错误，仅将明确无音源或试听判为不可播放。
func TestSongURLErrors(t *testing.T) {
	for _, tc := range []struct {
		name, body, message string
		unavailable         bool
	}{
		{"risk", `{"code":-460,"message":"network risk"}`, "-460", false},
		{"login", `{"code":301,"msg":"login required"}`, "login required", false},
		{"empty", `{"code":200,"data":[]}`, "", false},
		{"unavailable", `{"code":200,"data":[{"url":null}]}`, "", true},
		{"preview", `{"code":200,"data":[{"url":"https://audio.invalid/song.mp3","freeTrialInfo":{"start":0,"end":30}}]}`, "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			old := http.DefaultTransport
			http.DefaultTransport = songTransport(func(r *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(tc.body))}, nil
			})
			defer func() { http.DefaultTransport = old }()
			st, err := SongURL(1, "exhigh")
			if st != nil || err == nil {
				t.Fatalf("stream=%v err=%v", st, err)
			}
			if errors.Is(err, ErrUnavailable) != tc.unavailable {
				t.Fatalf("unavailable=%v err=%v", tc.unavailable, err)
			}
			if tc.message != "" && !strings.Contains(err.Error(), tc.message) {
				t.Fatalf("lost API error: %v", err)
			}
		})
	}
}

type songTransport func(*http.Request) (*http.Response, error)

func (f songTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
