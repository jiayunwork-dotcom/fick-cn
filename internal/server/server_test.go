package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestHealthAndStepClosedRod(t *testing.T) {
	root := testdataRoot(t)
	h := New(filepath.Join(root, "web"), filepath.Join(root, "example", "closed-rod.json"))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("health %d", rr.Code)
	}
	body, err := os.ReadFile(filepath.Join(root, "example", "closed-rod.json"))
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/step", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("step %d %s", rr.Code, rr.Body.String())
	}
	var got StepResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Nodes != 41 {
		t.Fatalf("nodes %d", got.Nodes)
	}
	rel := (got.MassFinal - got.Mass0) / got.Mass0
	if rel > 1e-8 || rel < -1e-8 {
		t.Fatalf("closed-rod mass drifted %g -> %g", got.Mass0, got.MassFinal)
	}
}

func testdataRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "example", "closed-rod.json")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	t.Fatal("cannot find example/closed-rod.json")
	return ""
}
