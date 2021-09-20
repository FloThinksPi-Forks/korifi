package apis_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.cloudfoundry.org/cf-k8s-api/apis"
	"code.cloudfoundry.org/cf-k8s-api/apis/fake"
	"code.cloudfoundry.org/cf-k8s-api/repositories"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"k8s.io/client-go/rest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestRoute(t *testing.T) {
	spec.Run(t, "RouteHandler", testRouteHandler, spec.Report(report.Terminal{}))
}

func testRouteHandler(t *testing.T, when spec.G, it spec.S) {
	g := NewWithT(t)

	var (
		rr            *httptest.ResponseRecorder
		routeRepo     *fake.CFRouteRepository
		domainRepo    *fake.CFDomainRepository
		clientBuilder *fake.ClientBuilder
		routeHandler  *apis.RouteHandler
		req           *http.Request
	)

	const (
		expectedRouteGUID  = "test-route-guid"
		expectedDomainGUID = "test-domain-guid"
	)

	it.Before(func() {
		rr = httptest.NewRecorder()
		routeRepo = new(fake.CFRouteRepository)
		domainRepo = new(fake.CFDomainRepository)
		clientBuilder = new(fake.ClientBuilder)

		routeRepo.FetchRouteReturns(repositories.RouteRecord{
			GUID:      expectedRouteGUID,
			SpaceGUID: "test-space-guid",
			DomainRef: repositories.DomainRecord{
				GUID: expectedDomainGUID,
			},
			Host:     "test-route-name",
			Protocol: "http",
		}, nil)

		domainRepo.FetchDomainReturns(repositories.DomainRecord{
			GUID: expectedDomainGUID,
			Name: "example.org",
		}, nil)

		routeHandler = &apis.RouteHandler{
			ServerURL:   defaultServerURL,
			RouteRepo:   routeRepo,
			DomainRepo:  domainRepo,
			BuildClient: clientBuilder.Spy,
			Logger:      logf.Log.WithName("TestRouteHandler"),
			K8sConfig:   &rest.Config{}, // required for k8s client (transitive dependency from route repo)
		}

		var err error
		req, err = http.NewRequest("GET", fmt.Sprintf("/v3/routes/%s", expectedRouteGUID), nil)
		g.Expect(err).NotTo(HaveOccurred())
	})

	when("the GET /v3/routes/:guid endpoint returns successfully", func() {
		it.Before(func() {
			http.HandlerFunc(routeHandler.RouteGetHandler).ServeHTTP(rr, req)
		})

		it("returns status 200 OK", func() {
			httpStatus := rr.Code
			g.Expect(httpStatus).Should(Equal(http.StatusOK), "Matching HTTP response code:")
		})

		it("returns Content-Type as JSON in header", func() {
			contentTypeHeader := rr.Header().Get("Content-Type")
			g.Expect(contentTypeHeader).Should(Equal(jsonHeader), "Matching Content-Type header:")
		})

		it("returns the Route in the response", func() {
			expectedBody := `{
				"guid":     "test-route-guid",
				"port": null,
				"path": "",
				"protocol": "http",
				"host":     "test-route-name",
				"url":      "test-route-name.example.org",
				"destinations": null,
				"relationships": {
					"space": {
						"data": {
							"guid": "test-space-guid"
						}
					},
					"domain": {
						"data": {
							"guid": "test-domain-guid"
						}
					}
				},
				"metadata": {
					"labels": {},
					"annotations": {}
				},
				"links": {
					"self":{
						"href": "https://api.example.org/v3/routes/test-route-guid"
					},
					"space":{
						"href": "https://api.example.org/v3/spaces/test-space-guid"
					},
					"domain":{
						"href": "https://api.example.org/v3/domains/test-domain-guid"
					},
					"destinations":{
						"href": "https://api.example.org/v3/routes/test-route-guid/destinations"
					}
				}
			}`

			g.Expect(rr.Body.String()).Should(MatchJSON(expectedBody), "Response body matches response:")
		})

		// The path parsing logic that extracts the route GUID relies on integration
		// with the mux router, which we don't use in test. We'll need to find a way to
		// get this test passing eventually or move this test to an integration-style test
		it.Pend("fetches the correct route and domain", func() {
			g.Expect(routeRepo.FetchRouteCallCount()).To(Equal(1))
			_, _, actualRouteGUID := routeRepo.FetchRouteArgsForCall(0)
			g.Expect(actualRouteGUID).To(Equal(expectedRouteGUID))

			g.Expect(domainRepo.FetchDomainCallCount()).To(Equal(1))
			_, _, actualDomainGUID := domainRepo.FetchDomainArgsForCall(0)
			g.Expect(actualDomainGUID).To(Equal(expectedDomainGUID))
		})
	})

	when("the route cannot be found", func() {
		it.Before(func() {
			routeRepo.FetchRouteReturns(repositories.RouteRecord{}, repositories.NotFoundError{Err: errors.New("not found")})

			http.HandlerFunc(routeHandler.RouteGetHandler).ServeHTTP(rr, req)
		})

		it("returns status 404 NotFound", func() {
			g.Expect(rr.Code).Should(Equal(http.StatusNotFound), "Matching HTTP response code:")
		})

		it("returns Content-Type as JSON in header", func() {
			contentTypeHeader := rr.Header().Get("Content-Type")
			g.Expect(contentTypeHeader).Should(Equal(jsonHeader), "Matching Content-Type header:")
		})

		it("returns a CF API formatted Error response", func() {
			g.Expect(rr.Body.String()).Should(MatchJSON(`{
				"errors": [
					{
						"code": 10010,
						"title": "CF-ResourceNotFound",
						"detail":"Route not found"
					}
				]
			}`), "Response body matches response:")
		})
	})

	when("the route's domain cannot be found", func() {
		it.Before(func() {
			domainRepo.FetchDomainReturns(repositories.DomainRecord{}, repositories.NotFoundError{Err: errors.New("not found")})

			http.HandlerFunc(routeHandler.RouteGetHandler).ServeHTTP(rr, req)
		})

		it("returns status 500 InternalServerError", func() {
			g.Expect(rr.Code).Should(Equal(http.StatusInternalServerError), "Matching HTTP response code:")
		})

		it("returns Content-Type as JSON in header", func() {
			contentTypeHeader := rr.Header().Get("Content-Type")
			g.Expect(contentTypeHeader).Should(Equal(jsonHeader), "Matching Content-Type header:")
		})

		it("returns a CF API formatted Error response", func() {
			g.Expect(rr.Body.String()).Should(MatchJSON(`{
				"errors": [
					{
						"title": "UnknownError",
						"detail": "An unknown error occurred.",
						"code": 10001
					}
				]
			}`), "Response body matches response:")
		})
	})

	when("there is some other error fetching the route", func() {
		it.Before(func() {
			routeRepo.FetchRouteReturns(repositories.RouteRecord{}, errors.New("unknown!"))

			http.HandlerFunc(routeHandler.RouteGetHandler).ServeHTTP(rr, req)
		})

		it("returns status 500 InternalServerError", func() {
			g.Expect(rr.Code).Should(Equal(http.StatusInternalServerError), "Matching HTTP response code:")
		})

		it("returns Content-Type as JSON in header", func() {
			contentTypeHeader := rr.Header().Get("Content-Type")
			g.Expect(contentTypeHeader).Should(Equal(jsonHeader), "Matching Content-Type header:")
		})

		it("returns a CF API formatted Error response", func() {
			g.Expect(rr.Body.String()).Should(MatchJSON(`{
				"errors": [
					{
						"title": "UnknownError",
						"detail": "An unknown error occurred.",
						"code": 10001
					}
				]
			}`), "Response body matches response:")
		})
	})
}
