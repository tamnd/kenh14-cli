package kenh14

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestClient(srv *httptest.Server) *Client {
	cfg := DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0
	cfg.Retries = 0
	cfg.Timeout = 5 * time.Second
	return NewClientWithConfig(cfg)
}

func sampleRSS(n int) string {
	items := ""
	for i := 0; i < n; i++ {
		items += fmt.Sprintf(`
		<item>
			<title>Bài viết giải trí số %d</title>
			<link>https://kenh14.vn/sao-viet/bai-viet-so-%d-1900%02d.chn</link>
			<description>Mô tả ngắn về bài viết %d</description>
			<pubDate>Sun, 14 Jun 2026 08:00:00 +0700</pubDate>
			<author>PV Kenh14</author>
			<thumb>https://cdn.kenh14.vn/img%d.jpg</thumb>
		</item>`, i+1, i+1, i+1, i+1, i+1)
	}
	return `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<title>Kenh14 RSS</title>
		<link>https://kenh14.vn</link>` + items + `
	</channel>
</rss>`
}

func TestGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("no User-Agent header")
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	body, err := c.Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "ok" {
		t.Errorf("body = %q, want ok", body)
	}
}

func TestGetRetriesOn503(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("recovered"))
	}))
	defer srv.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0
	cfg.Retries = 5
	cfg.Timeout = 5 * time.Second
	c := NewClientWithConfig(cfg)

	start := time.Now()
	body, err := c.Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "recovered" {
		t.Errorf("body = %q", body)
	}
	if hits != 3 {
		t.Errorf("hits = %d, want 3", hits)
	}
	if time.Since(start) < 500*time.Millisecond {
		t.Error("no backoff between retries")
	}
}

func TestLatestArticles(t *testing.T) {
	feed := sampleRSS(5)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(feed))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	articles, err := c.LatestArticles(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(articles) != 5 {
		t.Fatalf("len = %d, want 5", len(articles))
	}
	a := articles[0]
	if a.ID == "" {
		t.Error("ID empty")
	}
	if a.Category != "home" {
		t.Errorf("category = %q, want home", a.Category)
	}
	if a.Thumbnail == "" {
		t.Error("thumbnail empty (expected <thumb> tag)")
	}
}

func TestCategoryArticles(t *testing.T) {
	feed := sampleRSS(3)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(feed))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	articles, err := c.CategoryArticles(context.Background(), "sao-viet", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(articles) != 3 {
		t.Fatalf("len = %d, want 3", len(articles))
	}
	if articles[0].Category != "sao-viet" {
		t.Errorf("category = %q, want sao-viet", articles[0].Category)
	}
}

func TestCategoryLimit(t *testing.T) {
	feed := sampleRSS(10)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(feed))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	articles, err := c.CategoryArticles(context.Background(), "phim", 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(articles) != 4 {
		t.Fatalf("len = %d, want 4", len(articles))
	}
}

func TestExtractArticleID(t *testing.T) {
	cases := []struct{ url, want string }{
		{"https://kenh14.vn/sao-viet/ca-si-nao-do-190001.chn", "190001"},
		{"https://kenh14.vn/phim/phim-hay-1234567.chn", "1234567"},
		{"https://kenh14.vn/no-id-here.chn", ""},
	}
	for _, tc := range cases {
		got := extractArticleID(tc.url)
		if got != tc.want {
			t.Errorf("extractArticleID(%q) = %q, want %q", tc.url, got, tc.want)
		}
	}
}

func TestListCategories(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer srv.Close()

	c := newTestClient(srv)
	cats := c.ListCategories()
	if len(cats) == 0 {
		t.Fatal("empty categories")
	}
	if cats[0].Slug != "home" {
		t.Errorf("first slug = %q, want home", cats[0].Slug)
	}
}

func TestGetHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Get(context.Background(), srv.URL)
	if err == nil {
		t.Error("want error on 404")
	}
}
