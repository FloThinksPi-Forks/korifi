package e2e_test

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/go-resty/resty/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	rbacv1 "k8s.io/api/rbac/v1"
)

var _ = Describe("Orgs", func() {
	var resp *resty.Response

	Describe("create", func() {
		var (
			result    resource
			client    *resty.Client
			resultErr cfErrs
			orgName   string
		)

		BeforeEach(func() {
			client = adminClient
			orgName = generateGUID("my-org")
		})

		AfterEach(func() {
			deleteOrg(result.GUID)
		})

		JustBeforeEach(func() {
			var err error
			resp, err = client.R().
				SetBody(resource{Name: orgName}).
				SetError(&resultErr).
				SetResult(&result).
				Post("/v3/organizations")
			Expect(err).NotTo(HaveOccurred())
		})

		It("succeeds", func() {
			Expect(resp.StatusCode()).To(Equal(http.StatusCreated))
			Expect(result.Name).To(Equal(orgName))
			Expect(result.GUID).NotTo(BeEmpty())
		})

		When("the org name already exists", func() {
			var duplOrgGUID string

			BeforeEach(func() {
				duplOrgGUID = createOrg(orgName)
			})

			AfterEach(func() {
				deleteOrg(duplOrgGUID)
			})

			It("returns an unprocessable entity error", func() {
				Expect(resp.StatusCode()).To(Equal(http.StatusUnprocessableEntity))
				Expect(resultErr.Errors).To(HaveLen(1))
				Expect(resultErr.Errors[0].Code).To(BeNumerically("==", 10008))
				Expect(resultErr.Errors[0].Detail).To(MatchRegexp(fmt.Sprintf(`Organization '%s' already exists.`, orgName)))
				Expect(resultErr.Errors[0].Title).To(Equal("CF-UnprocessableEntity"))
			})
		})

		When("not admin", func() {
			BeforeEach(func() {
				client = tokenClient
			})

			It("returns a forbidden error", func() {
				Expect(resp.StatusCode()).To(Equal(http.StatusForbidden))
			})
		})
	})

	Describe("list", func() {
		var (
			org1Name, org2Name, org3Name, org4Name string
			org1GUID, org2GUID, org3GUID, org4GUID string
			result                                 resourceList
			query                                  map[string]string
		)

		BeforeEach(func() {
			var wg sync.WaitGroup
			errChan := make(chan error, 4)
			query = make(map[string]string)

			org1Name = generateGUID("org1")
			org2Name = generateGUID("org2")
			org3Name = generateGUID("org3")
			org4Name = generateGUID("org4")

			wg.Add(4)
			asyncCreateOrg(org1Name, &org1GUID, &wg, errChan)
			asyncCreateOrg(org2Name, &org2GUID, &wg, errChan)
			asyncCreateOrg(org3Name, &org3GUID, &wg, errChan)
			asyncCreateOrg(org4Name, &org4GUID, &wg, errChan)
			wg.Wait()

			var err error
			Expect(errChan).ToNot(Receive(&err), func() string { return fmt.Sprintf("unexpected error occurred while creating orgs: %v", err) })
			close(errChan)

			createOrgRole("organization_manager", rbacv1.ServiceAccountKind, serviceAccountName, org1GUID)
			createOrgRole("organization_manager", rbacv1.ServiceAccountKind, serviceAccountName, org2GUID)
			createOrgRole("organization_manager", rbacv1.ServiceAccountKind, serviceAccountName, org3GUID)
		})

		AfterEach(func() {
			var wg sync.WaitGroup
			wg.Add(4)
			for _, id := range []string{org1GUID, org2GUID, org3GUID, org4GUID} {
				asyncDeleteOrg(id, &wg)
			}
			wg.Wait()
		})

		JustBeforeEach(func() {
			var err error
			resp, err = tokenClient.R().
				SetQueryParams(query).
				SetResult(&result).
				Get("/v3/organizations")
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns orgs that the service account has a role in", func() {
			Expect(resp.StatusCode()).To(Equal(http.StatusOK))
			Expect(result.Resources).To(ConsistOf(
				MatchFields(IgnoreExtras, Fields{"Name": Equal(org1Name)}),
				MatchFields(IgnoreExtras, Fields{"Name": Equal(org2Name)}),
				MatchFields(IgnoreExtras, Fields{"Name": Equal(org3Name)}),
			))
			Expect(result.Resources).ToNot(ContainElement(
				MatchFields(IgnoreExtras, Fields{"Name": Equal(org4Name)}),
			))
		})

		When("org names are filtered", func() {
			BeforeEach(func() {
				query = map[string]string{
					"names": org1Name + "," + org3Name,
				}
			})

			It("returns orgs 1 & 3", func() {
				Expect(result.Resources).To(ConsistOf(
					MatchFields(IgnoreExtras, Fields{"Name": Equal(org1Name)}),
					MatchFields(IgnoreExtras, Fields{"Name": Equal(org3Name)}),
				))
				Expect(result.Resources).ToNot(ContainElement(
					MatchFields(IgnoreExtras, Fields{"Name": Equal(org2Name)}),
				))
			})
		})
	})
})
