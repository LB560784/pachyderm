//go:build unit_test

package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	pachdhttp "github.com/pachyderm/pachyderm/v2/src/server/http"
)

func TestRouting(t *testing.T) {
	testData := []struct {
		name     string
		method   string
		url      string
		wantCode int
	}{
		{
			name:     "not found",
			method:   "GET",
			url:      "http://pachyderm.example.com/",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "health",
			method:   "GET",
			url:      "http://pachyderm.example.com/healthz",
			wantCode: http.StatusOK,
		},
		{
			name:     "CreatePipelineRequest JSON schema",
			method:   "GET",
			url:      "http://pachyderm.example.com/jsonschema/pps_v2/CreatePipelineRequest.schema.json",
			wantCode: http.StatusOK,
		},
	}
	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			ctx := pctx.TestContext(t)

			s := pachdhttp.New(ctx, 0, func(ctx context.Context) *client.APIClient {
				env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
				client := env.PachClient
				return client
			})
			log.AddLoggerToHTTPServer(ctx, test.name, s.Server)

			req := httptest.NewRequest(test.method, test.url, nil)
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()

			s.Server.Handler.ServeHTTP(rec, req)
			if got, want := rec.Code, test.wantCode; got != want {
				t.Errorf("response code:\n  got: %v\n want: %v", got, want)
			}
		})
	}
}
