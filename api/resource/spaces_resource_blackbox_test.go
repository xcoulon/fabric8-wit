package resource_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fabric8-services/fabric8-wit/api"
	"github.com/fabric8-services/fabric8-wit/gormapplication"
	"github.com/fabric8-services/fabric8-wit/gormtestsupport"
	"github.com/stretchr/testify/suite"
)

type SpacesResourceBlackBoxTestSuite struct {
	gormtestsupport.DBTestSuite
	clean func()
	ctx   context.Context
}

func TestSpacesResource(t *testing.T) {
	suite.Run(t, &SpacesResourceBlackBoxTestSuite{DBTestSuite: gormtestsupport.NewDBTestSuite("../../config.yaml")})
}

func (s *SpacesResourceBlackBoxTestSuite) TestShowSpaceOK() {
	r, err := http.NewRequest(http.MethodGet, "/api/spaces/2e0698d8-753e-4cef-bb7c-f027634824a2", nil)
	if err != nil {
		s.T().Fatal(err)
	}
	// r.Header.Set(headerAccept, jsonapi.MediaType)

	rr := httptest.NewRecorder()
	httpEngine := api.NewGinEngine(gormapplication.NewGormDB(s.DB), s.Configuration)
	httpEngine.ServeHTTP(rr, r)

	if e, a := http.StatusOK, rr.Code; e != a {
		s.T().Fatalf("Expected a status of %d, got %d", e, a)
	}
}