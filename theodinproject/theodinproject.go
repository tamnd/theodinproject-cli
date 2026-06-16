// Package theodinproject is the library behind the theodinproject CLI.
package theodinproject

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const DefaultUserAgent = "theodinproject-cli/dev (+https://github.com/tamnd/theodinproject-cli)"

type Config struct {
	BaseURL   string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
	UserAgent string
}

func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://www.theodinproject.com",
		Rate:      500 * time.Millisecond,
		Timeout:   30 * time.Second,
		Retries:   3,
		UserAgent: DefaultUserAgent,
	}
}

type Client struct {
	cfg  Config
	http *http.Client
	last time.Time
}

func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

var (
	pathRe   = regexp.MustCompile(`href="/paths/([\w-]+)"[^>]*>\s*<h2[^>]*>([\s\S]*?)</h2>`)
	h3Re     = regexp.MustCompile(`<h3[^>]*>([^<]+)</h3>`)
	lessonRe = regexp.MustCompile(`/lessons/([\w-]+)"[^>]*>[\s\S]*?<p class="text-gray[^"]*">([^<]+)</p>`)
	tagRe    = regexp.MustCompile(`<[^>]+>`)
)

// Paths fetches the /paths page and returns all learning paths.
func (c *Client) Paths(ctx context.Context) ([]*Path, error) {
	body, err := c.get(ctx, c.cfg.BaseURL+"/paths")
	if err != nil {
		return nil, err
	}
	html := string(body)

	var paths []*Path
	rank := 0
	for _, m := range pathRe.FindAllStringSubmatch(html, -1) {
		rank++
		title := strings.TrimSpace(tagRe.ReplaceAllString(m[2], ""))
		paths = append(paths, &Path{
			Rank:  rank,
			Slug:  m[1],
			Title: title,
			URL:   c.cfg.BaseURL + "/paths/" + m[1],
		})
	}
	return paths, nil
}

// Lessons fetches a path page and returns all lessons grouped by course.
func (c *Client) Lessons(ctx context.Context, pathSlug string) ([]*Lesson, error) {
	body, err := c.get(ctx, c.cfg.BaseURL+"/paths/"+pathSlug)
	if err != nil {
		return nil, err
	}
	html := string(body)

	sections := splitByCourse(html)

	var lessons []*Lesson
	rank := 0
	for _, sec := range sections {
		course := sec.course
		for _, m := range lessonRe.FindAllStringSubmatch(sec.html, -1) {
			rank++
			slug := m[1]
			title := strings.TrimSpace(m[2])
			lessons = append(lessons, &Lesson{
				Rank:   rank,
				Slug:   slug,
				Course: course,
				Title:  title,
				URL:    c.cfg.BaseURL + "/lessons/" + slug,
			})
		}
	}
	return lessons, nil
}

type courseSection struct {
	course string
	html   string
}

func splitByCourse(html string) []courseSection {
	h3Matches := h3Re.FindAllStringSubmatchIndex(html, -1)
	var sections []courseSection
	curCourse := ""
	curStart := 0

	for _, m := range h3Matches {
		h3Text := strings.TrimSpace(html[m[2]:m[3]])
		if isFooterH3(h3Text) {
			continue
		}
		if curCourse != "" {
			end := m[0]
			sections = append(sections, courseSection{course: curCourse, html: html[curStart:end]})
		}
		curCourse = h3Text
		curStart = m[1]
	}
	if curCourse != "" {
		sections = append(sections, courseSection{course: curCourse, html: html[curStart:]})
	}
	return sections
}

var footerH3s = map[string]bool{
	"About us": true, "Support": true, "Guides": true, "Legal": true,
}

func isFooterH3(s string) bool { return footerH3s[s] }

// Info returns site-level stats.
func (c *Client) Info(ctx context.Context) (*Info, error) {
	paths, err := c.Paths(ctx)
	if err != nil {
		return nil, err
	}
	return &Info{
		Site:   "www.theodinproject.com",
		Paths:  len(paths),
		Source: c.cfg.BaseURL,
	}, nil
}

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, url)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, url string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) pace() {
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}
