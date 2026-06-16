package theodinproject_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tamnd/theodinproject-cli/theodinproject"
)

func TestInfo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, fakePathsPage)
	}))
	defer ts.Close()

	cfg := theodinproject.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	c := theodinproject.NewClient(cfg)

	info, err := c.Info(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.Paths != 2 {
		t.Errorf("Paths = %d, want 2", info.Paths)
	}
	if info.Site == "" {
		t.Error("Site should not be empty")
	}
}
