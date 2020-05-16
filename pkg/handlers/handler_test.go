package handlers

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	// "go.uber.org/zap"
	// "go.uber.org/zap/zaptest"
)

// func TestHandler(t *testing.T) {
// 	logger := zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel))
// 	h := NewHandler(logger.Sugar(), nil)
// 	rec := httptest.NewRecorder()
// 	req := makeHookRequest(t, "testdata/push_hook.json")

// 	h.ServeHTTP(rec, req)
// }

func makeHookRequest(t *testing.T, fixture string) *http.Request {
	t.Helper()
	b, err := ioutil.ReadFile(fixture)
	if err != nil {
		t.Fatalf("failed to read %s: %s", fixture, err)
	}
	req := httptest.NewRequest("POST", "/", bytes.NewReader(b))
	req.Header.Add("Content-Type", "application/json")
	return req
}
