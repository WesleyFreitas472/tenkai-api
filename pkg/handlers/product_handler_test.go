package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	mockRepo "github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	mockSvc "github.com/softplan/tenkai-api/pkg/service/_helm/mocks"
	"github.com/softplan/tenkai-api/pkg/service/docker/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSplitSrvNameIfNeeded(t *testing.T) {
	assert.Equal(t, "repo/my-chart", splitSrvNameIfNeeded("repo/my-chart - 0.1.0"))
	assert.Equal(t, "repo/my-chart", splitSrvNameIfNeeded("repo/my-chart"))
}

func TestSplitChartVersion(t *testing.T) {
	assert.Equal(t, "0.1.0", splitChartVersion("repo/my-chart - 0.1.0"))
	assert.Equal(t, "", splitChartVersion("repo/my-chart"))
}

func TestSplitChartRepo(t *testing.T) {
	assert.Equal(t, "repo", splitChartRepo("repo/my-chart - 0.1.0"))
	assert.Equal(t, "", splitChartRepo("my-chart"))
}

func TestGetChartLatestVersion(t *testing.T) {
	appContext := AppContext{}

	var sr1 model.SearchResult
	sr1.Name = "repo/my-chart"
	sr1.ChartVersion = "0.1.0"
	sr1.AppVersion = "1.0.0"
	sr1.Description = "This is my chart"

	var results []model.SearchResult
	results = append(results, sr1)

	latestVersion := appContext.getChartLatestVersion("repo/my-chart - 0.1.0", results)
	assert.Equal(t, "", latestVersion, "Should not have a latest version")

	var sr2 model.SearchResult
	sr2.Name = "repo/my-chart"
	sr2.ChartVersion = "0.2.0"
	sr2.AppVersion = "1.0.0"
	sr2.Description = "This is my chart"
	results = append(results, sr2)

	latestVersion = appContext.getChartLatestVersion("repo/my-chart - 0.1.0", results)
	assert.Equal(t, "0.2.0", latestVersion, "Latest version should be 0.2.0")
}

func TestNewProduct(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("CreateProduct", getProductWithoutID()).Return(999, nil)
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("POST", "/products", payload(getProductWithoutID()))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newProduct)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "CreateProduct", 1)
	assert.Equal(t, http.StatusCreated, rr.Code, "Response should be Created")
}

func TestNewProduct_UnmarshalPayloadError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadError(t, "/products", appContext.newProduct)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestNewProduct_CreateProductError(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("CreateProduct", getProductWithoutID()).Return(0, errors.New("Error saving product"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("POST", "/products", payload(getProductWithoutID()))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newProduct)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestEditProduct(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("EditProduct", getProduct()).Return(nil)
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("POST", "/products/edit", payload(getProduct()))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editProduct)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "EditProduct", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok")
}

func TestEditProduct_UnmarshalPayloadError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadError(t, "/products/edit", appContext.editProduct)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestEditProduct_CreateProductError(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("EditProduct", getProduct()).Return(errors.New("Error saving product"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("POST", "/products/edit", payload(getProduct()))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editProduct)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestDeleteProduct(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("DeleteProduct", 999).Return(nil)
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("DELETE", "/products/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/products/{id}", appContext.deleteProduct).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "DeleteProduct", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
}

func TestDeleteProduct_Error(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("DeleteProduct", 999).Return(errors.New("Error deleting product"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("DELETE", "/products/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/products/{id}", appContext.deleteProduct).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "DeleteProduct", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestListProducts(t *testing.T) {
	result := &model.ProductRequestReponse{}
	result.List = append(result.List, getProduct())

	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProducts").Return(result.List, nil)
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("GET", "/products", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listProducts)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProducts", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok")

	response := string(rr.Body.Bytes())
	assert.Contains(t, response, `{"list":[{"ID":999,`)
	assert.Contains(t, response, `"name":"my-product",`)
	assert.Contains(t, response, `"validateReleases":true}]}`)
}

func TestListProducts_Error(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProducts").Return(nil, errors.New("Error listing product"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("GET", "/products", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listProducts)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProducts", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestNewProductVersion(t *testing.T) {
	appContext := AppContext{}
	childs := getProductVersionSvcReqResp()

	pv := getProductVersionWithoutID(0)
	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("CreateProductVersionCopying", mock.Anything).Return(999, nil)
	mockProductDAO.On("ListProductsVersionServices", 999).Return(childs.List, nil)
	appContext.Repositories.ProductDAO = mockProductDAO

	webHooks := make([]model.WebHook, 0)
	webHooks = append(webHooks, mockWebHook())
	mockWebHookDAO := &mockRepo.WebHookDAOInterface{}
	mockWebHookDAO.On("ListWebHooksByEnvAndType", -1, "HOOK_NEW_RELEASE").
		Return(webHooks, nil)
	appContext.Repositories.WebHookDAO = mockWebHookDAO

	var product model.Product
	product.ID = 999
	product.Name = "My Product"
	mockProductDAO.On("FindProductByID", 999).Return(product, nil)
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("POST", "/productVersions", payload(pv))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newProductVersion)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "CreateProductVersionCopying", 1)
	assert.Equal(t, http.StatusCreated, rr.Code, "Response should be Created")
}

func TestNewProductVersion_UnmarshalPayloadError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadError(t, "/productVersions", appContext.newProductVersion)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestNewProductVersion_Error(t *testing.T) {
	appContext := AppContext{}

	pv := getProductVersionWithoutID(0)
	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("CreateProductVersionCopying", mock.Anything).Return(0, errors.New("Some error"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("POST", "/productVersions", payload(pv))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newProductVersion)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "CreateProductVersionCopying", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500")
}

func TestEditProductVersion(t *testing.T) {
	appContext := AppContext{}

	pv := getProductVersionWithoutID(0)
	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("EditProductVersion", mock.Anything).Return(nil)
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("POST", "/productVersions/edit", payload(pv))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editProductVersion)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "EditProductVersion", 1)
	assert.Equal(t, http.StatusCreated, rr.Code, "Response should be Created")
}

func TestEditProductVersion_UnmarshalPayloadError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadError(t, "/productVersions/edit",
		appContext.editProductVersion)
	assert.Equal(t, http.StatusInternalServerError, rr.Code,
		"Response should be 500.")
}

func TestEditProductVersion_Error(t *testing.T) {
	appContext := AppContext{}

	pv := getProductVersionWithoutID(0)
	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("EditProductVersion", mock.Anything).
		Return(errors.New("Some error"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("POST", "/productVersions/edit", payload(pv))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editProductVersion)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "EditProductVersion", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code,
		"Response should be 500")
}

func TestDeleteProductVersion(t *testing.T) {
	appContext := AppContext{}

	childs := getProductVersionSvcReqResp()

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProductsVersionServices", 999).Return(childs.List, nil)
	mockProductDAO.On("DeleteProductVersionService", 888).Return(nil)
	mockProductDAO.On("DeleteProductVersion", 999).Return(nil)
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("DELETE", "/productVersions/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/productVersions/{id}", appContext.deleteProductVersion).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductsVersionServices", 1)
	mockProductDAO.AssertNumberOfCalls(t, "DeleteProductVersionService", 1)
	mockProductDAO.AssertNumberOfCalls(t, "DeleteProductVersion", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok.")
}

func TestDeleteProductVersion_ListProductsVersionServicesError(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProductsVersionServices", 999).Return(nil, errors.New("Some error"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("DELETE", "/productVersions/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/productVersions/{id}", appContext.deleteProductVersion).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductsVersionServices", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestDeleteProductVersion_DeleteProductVersionServiceError(t *testing.T) {
	appContext := AppContext{}

	childs := getProductVersionSvcReqResp()

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProductsVersionServices", 999).Return(childs.List, nil)
	mockProductDAO.On("DeleteProductVersionService", 888).Return(errors.New("Some error"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("DELETE", "/productVersions/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/productVersions/{id}", appContext.deleteProductVersion).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductsVersionServices", 1)
	mockProductDAO.AssertNumberOfCalls(t, "DeleteProductVersionService", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestDeleteProductVersion_DeleteProductVersionError(t *testing.T) {
	appContext := AppContext{}

	childs := getProductVersionSvcReqResp()

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProductsVersionServices", 999).Return(childs.List, nil)
	mockProductDAO.On("DeleteProductVersionService", 888).Return(nil)
	mockProductDAO.On("DeleteProductVersion", 999).Return(errors.New("Some error"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("DELETE", "/productVersions/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/productVersions/{id}", appContext.deleteProductVersion).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductsVersionServices", 1)
	mockProductDAO.AssertNumberOfCalls(t, "DeleteProductVersionService", 1)
	mockProductDAO.AssertNumberOfCalls(t, "DeleteProductVersion", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestListProductVersions(t *testing.T) {
	appContext := AppContext{}

	result := getProductVersionReqResp()

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProductsVersions", 777).Return(result.List, nil)
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("GET", "/productVersions/?productId=777", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listProductVersions)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductsVersions", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok.")

	response := string(rr.Body.Bytes())
	assert.Contains(t, response, `{"list":[{"ID":777,`)
	assert.Contains(t, response, `"version":"19.0.1-0",`)
	assert.Contains(t, response, `"baseRelease":0,`)
	assert.Contains(t, response, `"locked":false,"hotFix":false}]}`)
}

func TestListProductVersions_QueryError(t *testing.T) {
	appContext := AppContext{}

	req, err := http.NewRequest("GET", "/productVersions/?foo=bar", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listProductVersions)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestListProductVersions_ListProductsVersionsError(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProductsVersions", 777).Return(nil, errors.New("some error"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("GET", "/productVersions/?productId=777", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listProductVersions)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductsVersions", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestLockProductVersion(t *testing.T) {
	appContext := AppContext{}

	var p model.ProductVersion
	p.ID = uint(777)
	p.Version = "19.0.1-0"

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProductVersionsByID", mock.Anything).Return(&p, nil)
	mockProductDAO.On("EditProductVersion", mock.Anything).Return(nil)
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("GET", "/productVersions/lock/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/productVersions/lock/{id}", appContext.lockProductVersion).Methods("GET")
	r.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductVersionsByID", 1)
	mockProductDAO.AssertNumberOfCalls(t, "EditProductVersion", 1)

	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok.")
}

func TestLockProductVersion_Unauthorized(t *testing.T) {
	appContext := AppContext{}

	req, err := http.NewRequest("GET", "/productVersions/lock/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/productVersions/lock/{id}", appContext.lockProductVersion).Methods("GET")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Response should be 401.")
}

func TestLockProductVersion_StringConvError(t *testing.T) {
	appContext := AppContext{}

	req, err := http.NewRequest("GET", "/productVersions/lock/qwert", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/productVersions/lock/{id}", appContext.lockProductVersion).Methods("GET")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestLockProductVersion_ListProductVersionsByIDError(t *testing.T) {
	appContext := AppContext{}

	var p model.ProductVersion
	p.ID = uint(777)
	p.Version = "19.0.1-0"

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProductVersionsByID", mock.Anything).Return(nil, errors.New("some error"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("GET", "/productVersions/lock/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/productVersions/lock/{id}", appContext.lockProductVersion).Methods("GET")
	r.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductVersionsByID", 1)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestLockProductVersion_EditProductVersionError(t *testing.T) {
	appContext := AppContext{}

	var p model.ProductVersion
	p.ID = uint(777)
	p.Version = "19.0.1-0"

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProductVersionsByID", mock.Anything).Return(&p, nil)
	mockProductDAO.On("EditProductVersion", mock.Anything).Return(errors.New("some error"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("GET", "/productVersions/lock/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/productVersions/lock/{id}", appContext.lockProductVersion).Methods("GET")
	r.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductVersionsByID", 1)
	mockProductDAO.AssertNumberOfCalls(t, "EditProductVersion", 1)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestUnlockProductVersion(t *testing.T) {
	appContext := AppContext{}
	appContext.K8sConfigPath = "/tmp/"

	var p model.ProductVersion
	p.ID = uint(777)
	p.Version = "19.0.1-0"

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProductVersionsByID", mock.Anything).Return(&p, nil)
	mockProductDAO.On("EditProductVersion", mock.Anything).Return(nil)
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("GET", "/productVersions/unlock/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/productVersions/unlock/{id}", appContext.unlockProductVersion).Methods("GET")
	r.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductVersionsByID", 1)
	mockProductDAO.AssertNumberOfCalls(t, "EditProductVersion", 1)

	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok.")
}

func TestUnlockProductVersion_Unauthorized(t *testing.T) {
	appContext := AppContext{}

	req, err := http.NewRequest("GET", "/productVersions/unlock/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/productVersions/unlock/{id}", appContext.unlockProductVersion).Methods("GET")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Response should be 401.")
}

func TestUnlockProductVersion_StringConvError(t *testing.T) {
	appContext := AppContext{}
	appContext.K8sConfigPath = "/tmp/"

	req, err := http.NewRequest("GET", "/productVersions/unlock/qwert", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/productVersions/unlock/{id}", appContext.unlockProductVersion).Methods("GET")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestUnlockProductVersion_ListProductVersionsByIDError(t *testing.T) {
	appContext := AppContext{}

	var p model.ProductVersion
	p.ID = uint(777)
	p.Version = "19.0.1-0"

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProductVersionsByID", mock.Anything).Return(nil, errors.New("some error"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("GET", "/productVersions/unlock/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/productVersions/unlock/{id}", appContext.unlockProductVersion).Methods("GET")
	r.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductVersionsByID", 1)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestUnlockProductVersion_EditProductVersionError(t *testing.T) {
	appContext := AppContext{}

	var p model.ProductVersion
	p.ID = uint(777)
	p.Version = "19.0.1-0"

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProductVersionsByID", mock.Anything).Return(&p, nil)
	mockProductDAO.On("EditProductVersion", mock.Anything).Return(errors.New("some error"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("GET", "/productVersions/unlock/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/productVersions/unlock/{id}", appContext.unlockProductVersion).Methods("GET")
	r.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductVersionsByID", 1)
	mockProductDAO.AssertNumberOfCalls(t, "EditProductVersion", 1)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestListProductVersionServices(t *testing.T) {
	appContext := AppContext{}

	var pvs []model.ProductVersionService
	pvs = append(pvs, getProductVersionSvc())

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProductsVersionServices", 999).Return(pvs, nil)

	pv := getProductVersionWithoutID(0)
	mockProductDAO.On("ListProductVersionsByID", mock.Anything).Return(&pv, nil)
	appContext.Repositories.ProductDAO = mockProductDAO

	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	data := getHelmSearchResult()
	mockHelmSvc.On("SearchCharts", mock.Anything, false).Return(&data)
	appContext.HelmServiceAPI = mockHelmSvc

	appContext.ChartImageCache.Store("repo/my-chart", "myrepo.com/my-chart")

	mockDockerSvc := mockGetDockerTagsWithDate(&appContext, getTagResponse("19.0.1-1"))

	req, err := http.NewRequest("GET", "/productVersionServices/?productVersionId=999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listProductVersionServices)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductsVersionServices", 1)
	mockHelmSvc.AssertNumberOfCalls(t, "SearchCharts", 1)
	mockDockerSvc.AssertNumberOfCalls(t, "GetDockerTagsWithDate", 1)

	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok.")

	response := string(rr.Body.Bytes())
	assert.Contains(t, response, `{"list":[{"ID":888,`)
	assert.Contains(t, response, `"productVersionId":999,`)
	assert.Contains(t, response, `"serviceName":"repo/my-chart - 0.1.0",`)
	assert.Contains(t, response, `"dockerImageTag":"19.0.1-0"`)
	assert.Contains(t, response, `"latestVersion":"19.0.1-1",`)
	assert.Contains(t, response, `"chartLatestVersion":"1.0",`)
	assert.Contains(t, response, `"notes":""}]}`)

}

func TestListProductVersionServices_QueryError(t *testing.T) {
	appContext := AppContext{}

	req, err := http.NewRequest("GET", "/productVersionServices/?foo=bar", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listProductVersionServices)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestListProductVersionServices_ListProductVersionsByIDError(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	pv := getProductVersionWithoutID(0)
	mockProductDAO.On("ListProductVersionsByID", mock.Anything).Return(&pv, errors.New("some error"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("GET", "/productVersionServices/?productVersionId=999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listProductVersionServices)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductVersionsByID", 1)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be Ok.")
}

func TestListProductVersionServices_ListProdVerSvcError(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	pv := getProductVersionWithoutID(0)
	mockProductDAO.On("ListProductVersionsByID", mock.Anything).Return(&pv, nil)
	mockProductDAO.On("ListProductsVersionServices", 999).Return(nil, errors.New("some error"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("GET", "/productVersionServices/?productVersionId=999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listProductVersionServices)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductsVersionServices", 1)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be Ok.")
}

func TestNewProductVersionService(t *testing.T) {
	appContext := AppContext{}

	pvs := getProductVersionSvc()
	mockProductDAO := &mockRepo.ProductDAOInterface{}

	pv := getProductVersionWithoutID(0)
	pv.ID = 999

	mockProductDAO.On("ListProductVersionsByID", 999).Return(&pv, nil)
	mockProductDAO.On("CreateProductVersionService", pvs).Return(888, nil)
	mockProductDAO.On("FindProductByID", 999).Return(getProduct(), nil)

	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("POST", "/productVersionServices", payload(pvs))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newProductVersionService)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductVersionsByID", 1)
	mockProductDAO.AssertNumberOfCalls(t, "FindProductByID", 1)
	mockProductDAO.AssertNumberOfCalls(t, "CreateProductVersionService", 1)

	assert.Equal(t, http.StatusCreated, rr.Code, "Response should be Created.")
}

func TestNewProductVersionService_UnmarshalPayloadError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadError(t, "/productVersionServices", appContext.newProductVersionService)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestNewProductVersionService_Error1(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProductVersionsByID", 999).Return(nil, errors.New("some error"))
	appContext.Repositories.ProductDAO = mockProductDAO

	pvs := getProductVersionSvc()
	req, err := http.NewRequest("POST", "/productVersionServices", payload(pvs))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newProductVersionService)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductVersionsByID", 1)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestNewProductVersionService_Error2(t *testing.T) {
	appContext := AppContext{}

	pvs := getProductVersionSvc()
	mockProductDAO := &mockRepo.ProductDAOInterface{}

	pv := getProductVersionWithoutID(0)
	pv.ID = 999
	pv.Locked = true
	mockProductDAO.On("ListProductVersionsByID", 999).Return(&pv, nil)

	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("POST", "/productVersionServices", payload(pvs))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newProductVersionService)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductVersionsByID", 1)

	assert.Equal(t, http.StatusBadRequest, rr.Code, "Response should be 400.")
}

func TestNewProductVersionService_Error3(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}

	pv := getProductVersionWithoutID(0)
	pv.ID = 999
	mockProductDAO.On("ListProductVersionsByID", 999).Return(&pv, nil)

	mockProductDAO.On("CreateProductVersionService", mock.Anything).Return(0, errors.New("some error"))
	mockProductDAO.On("FindProductByID", 999).Return(getProduct(), nil)
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("POST", "/productVersionServices", payload(getProductVersionSvc()))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newProductVersionService)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductVersionsByID", 1)
	mockProductDAO.AssertNumberOfCalls(t, "FindProductByID", 1)
	mockProductDAO.AssertNumberOfCalls(t, "CreateProductVersionService", 1)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be Created.")
}

func TestNewProductVersionService_Error4(t *testing.T) {
	appContext := AppContext{}

	pvs := getProductVersionSvc()
	mockProductDAO := &mockRepo.ProductDAOInterface{}

	pv := getProductVersionWithoutID(0)
	pv.ID = 999
	pv.Version = "19.0.2-0"
	mockProductDAO.On("ListProductVersionsByID", 999).Return(&pv, nil)
	mockProductDAO.On("FindProductByID", 999).Return(getProduct(), nil)

	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("POST", "/productVersionServices", payload(pvs))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newProductVersionService)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductVersionsByID", 1)
	mockProductDAO.AssertNumberOfCalls(t, "FindProductByID", 1)

	assert.Equal(t, http.StatusBadRequest, rr.Code, "Response should be 400.")
}

func TestNewProductVersionService_ReleaseValidationDisabled(t *testing.T) {
	appContext := AppContext{}

	pvs := getProductVersionSvc()
	mockProductDAO := &mockRepo.ProductDAOInterface{}

	pv := getProductVersionWithoutID(0)
	pv.ID = 999
	pv.Version = "19.0.2-0"

	p := getProduct()
	p.ValidateReleases = false

	mockProductDAO.On("ListProductVersionsByID", 999).Return(&pv, nil)
	mockProductDAO.On("FindProductByID", 999).Return(p, nil)
	mockProductDAO.On("CreateProductVersionService", pvs).Return(888, nil)

	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("POST", "/productVersionServices", payload(pvs))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newProductVersionService)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductVersionsByID", 1)
	mockProductDAO.AssertNumberOfCalls(t, "FindProductByID", 1)
	mockProductDAO.AssertNumberOfCalls(t, "CreateProductVersionService", 1)

	assert.Equal(t, http.StatusCreated, rr.Code, "Response should be 201.")
}

func TestEditProductVersionService(t *testing.T) {
	appContext := AppContext{}

	pvs := getProductVersionSvc()
	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("EditProductVersionService", pvs).Return(nil)

	pv := getProductVersionWithoutID(0)
	pv.ID = 999
	mockProductDAO.On("ListProductVersionsByID", 999).Return(&pv, nil)

	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("POST", "/productVersionServices/edit", payload(pvs))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editProductVersionService)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductVersionsByID", 1)
	mockProductDAO.AssertNumberOfCalls(t, "EditProductVersionService", 1)

	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok.")
}

func TestEditProductVersionService_UnmarshalPayloadError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadError(t, "/productVersionServices/edit", appContext.editProductVersionService)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestEditProductVersionService_ListProductVersionsByIDError(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProductVersionsByID", 999).Return(nil, errors.New("some error"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("POST", "/productVersionServices/edit", payload(getProductVersionSvc()))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editProductVersionService)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductVersionsByID", 1)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestEditProductVersionService_ProductVersionLocked(t *testing.T) {
	appContext := AppContext{}

	pvs := getProductVersionSvc()
	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("EditProductVersionService", pvs).Return(nil)

	pv := getProductVersionWithoutID(0)
	pv.ID = 999
	pv.Locked = true
	mockProductDAO.On("ListProductVersionsByID", 999).Return(&pv, nil)

	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("POST", "/productVersionServices/edit", payload(pvs))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editProductVersionService)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductVersionsByID", 1)
	mockProductDAO.AssertNumberOfCalls(t, "EditProductVersionService", 1)

	assert.Equal(t, http.StatusBadRequest, rr.Code, "Response should be 400.")
}

func TestEditProductVersionService_EditProductVersionServiceError(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}

	pv := getProductVersionWithoutID(0)
	pv.ID = 999
	mockProductDAO.On("ListProductVersionsByID", 999).Return(&pv, nil)

	mockProductDAO.On("EditProductVersionService", mock.Anything).Return(errors.New("some error"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("POST", "/productVersionServices/edit", payload(getProductVersionSvc()))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editProductVersionService)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductVersionsByID", 1)
	mockProductDAO.AssertNumberOfCalls(t, "EditProductVersionService", 1)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestDeleteProductVersionService(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}

	pvs := getProductVersionSvc()
	mockProductDAO.On("ListProductVersionsServiceByID", 888).Return(&pvs, nil)

	pv := getProductVersionWithoutID(0)
	pv.ID = 999
	mockProductDAO.On("ListProductVersionsByID", 999).Return(&pv, nil)

	mockProductDAO.On("DeleteProductVersionService", 888).Return(nil)
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("DELETE", "/productVersionServices/888", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/productVersionServices/{id}", appContext.deleteProductVersionService).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductVersionsServiceByID", 1)
	mockProductDAO.AssertNumberOfCalls(t, "ListProductVersionsByID", 1)
	mockProductDAO.AssertNumberOfCalls(t, "DeleteProductVersionService", 1)

	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok.")
}

func TestDeleteProductVersionService_Error1(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProductVersionsServiceByID", 888).Return(nil, errors.New("some error"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("DELETE", "/productVersionServices/888", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/productVersionServices/{id}", appContext.deleteProductVersionService).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductVersionsServiceByID", 1)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestDeleteProductVersionService_Error2(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}

	pvs := getProductVersionSvc()
	mockProductDAO.On("ListProductVersionsServiceByID", 888).Return(&pvs, nil)

	mockProductDAO.On("ListProductVersionsByID", 999).Return(nil, errors.New("some error"))

	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("DELETE", "/productVersionServices/888", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/productVersionServices/{id}", appContext.deleteProductVersionService).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductVersionsServiceByID", 1)
	mockProductDAO.AssertNumberOfCalls(t, "ListProductVersionsByID", 1)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestDeleteProductVersionService_Error3(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}

	pvs := getProductVersionSvc()
	mockProductDAO.On("ListProductVersionsServiceByID", 888).Return(&pvs, nil)

	pv := getProductVersionWithoutID(0)
	pv.ID = 999
	mockProductDAO.On("ListProductVersionsByID", 999).Return(&pv, nil)

	mockProductDAO.On("DeleteProductVersionService", 888).Return(errors.New("some error"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("DELETE", "/productVersionServices/888", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/productVersionServices/{id}", appContext.deleteProductVersionService).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "DeleteProductVersionService", 1)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestDeleteProductVersionService_Locked(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}

	pvs := getProductVersionSvc()
	mockProductDAO.On("ListProductVersionsServiceByID", 888).Return(&pvs, nil)

	pv := getProductVersionWithoutID(0)
	pv.ID = 999
	pv.Locked = true
	mockProductDAO.On("ListProductVersionsByID", 999).Return(&pv, nil)

	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("DELETE", "/productVersionServices/888", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/productVersionServices/{id}", appContext.deleteProductVersionService).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductVersionsServiceByID", 1)
	mockProductDAO.AssertNumberOfCalls(t, "ListProductVersionsByID", 1)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestVerifyNewVersion_1(t *testing.T) {
	appContext := AppContext{}

	appContext.ChartImageCache.Store("repo/my-chart - 0.1.0", "myrepo.com/my-chart")

	mockDockerSvc := mockGetDockerTagsWithDate(&appContext, getTagResponse("19.0.2-0"))

	version, err := appContext.verifyNewVersion("repo/my-chart - 0.1.0", "19.0.1-0", "19.0.1-0", false)
	assert.NoError(t, err)
	assert.NotNil(t, version)

	assert.Equal(t, "", version)

	mockDockerSvc.AssertNumberOfCalls(t, "GetDockerTagsWithDate", 1)
}

func TestVerifyNewVersion_2(t *testing.T) {
	appContext := AppContext{}

	appContext.ChartImageCache.Store("repo/my-chart - 0.1.0", "myrepo.com/my-chart")

	mockDockerSvc := mockGetDockerTagsWithDate(&appContext, getTagResponse("19.0.1-1"))

	version, err := appContext.verifyNewVersion("repo/my-chart - 0.1.0", "19.0.1-0", "19.0.1-0", false)
	assert.NoError(t, err)
	assert.NotNil(t, version)

	assert.Equal(t, "19.0.1-1", version)

	mockDockerSvc.AssertNumberOfCalls(t, "GetDockerTagsWithDate", 1)
}

func TestVerifyNewVersion_3(t *testing.T) {
	appContext := AppContext{}

	appContext.ChartImageCache.Store("repo/my-chart - 0.1.0", "myrepo.com/my-chart")

	mockDockerSvc := mockGetDockerTagsWithDate(&appContext, getTagResponse("20.1.0-RC-2"))

	version, err := appContext.verifyNewVersion("repo/my-chart - 0.1.0", "20.1.0-RC-1", "20.1.0-RC-1", false)
	assert.NoError(t, err)
	assert.NotNil(t, version)

	assert.Equal(t, "20.1.0-RC-2", version)

	mockDockerSvc.AssertNumberOfCalls(t, "GetDockerTagsWithDate", 1)
}

func TestVerifyNewVersion_4(t *testing.T) {
	appContext := AppContext{}

	appContext.ChartImageCache.Store("repo/my-chart - 0.1.0", "myrepo.com/my-chart")

	mockDockerSvc := mockGetDockerTagsWithDate(&appContext, getTagResponse("20.1.0-0.1"))

	version, err := appContext.verifyNewVersion("repo/my-chart - 0.1.0", "20.1.0-0", "20.1.0-0", false)
	assert.NoError(t, err)
	assert.NotNil(t, version)

	assert.Equal(t, "", version)

	mockDockerSvc.AssertNumberOfCalls(t, "GetDockerTagsWithDate", 1)
}

func TestVerifyNewVersion_5(t *testing.T) {
	appContext := AppContext{}

	appContext.ChartImageCache.Store("repo/my-chart - 0.1.0", "myrepo.com/my-chart")

	latest := "20.2.1-RC-1"
	productVersion := "20.2.1-RC-0"
	svcCurrentVersion := "20.2.0-RC-9"

	mockDockerSvc := mockGetDockerTagsWithDate(&appContext, getTagResponse(latest))
	version, err := appContext.verifyNewVersion("repo/my-chart - 0.1.0", svcCurrentVersion, productVersion, false)
	assert.NoError(t, err)
	assert.NotNil(t, version)

	assert.Equal(t, latest, version)

	mockDockerSvc.AssertNumberOfCalls(t, "GetDockerTagsWithDate", 1)
}

func TestVerifyNewVersionHotfix_6(t *testing.T) {
	appContext := AppContext{}

	appContext.ChartImageCache.Store("repo/my-chart - 0.1.0", "myrepo.com/my-chart")

	latest := "20.1.1-15.6"
	productVersion := "20.1.1-1.1"
	svcCurrentVersion := "20.1.1-15.5"
	hotFix := true

	mockDockerSvc := mockGetDockerTagsWithDate(&appContext, getTagResponse(latest))
	version, err := appContext.verifyNewVersion("repo/my-chart - 0.1.0", svcCurrentVersion, productVersion, hotFix)
	assert.NoError(t, err)
	assert.NotNil(t, version)

	assert.Equal(t, latest, version)

	mockDockerSvc.AssertNumberOfCalls(t, "GetDockerTagsWithDate", 1)
}

func TestVerifyNewVersionHotfix_7(t *testing.T) {
	appContext := AppContext{}

	appContext.ChartImageCache.Store("repo/my-chart - 0.1.0", "myrepo.com/my-chart")

	latest := "20.1.1-1.2"
	productVersion := "20.1.1-1.1"
	svcCurrentVersion := "20.1.1-15.5"
	hotFix := true

	mockDockerSvc := mockGetDockerTagsWithDate(&appContext, getTagResponse(latest))
	version, err := appContext.verifyNewVersion("repo/my-chart - 0.1.0", svcCurrentVersion, productVersion, hotFix)
	assert.NoError(t, err)
	assert.NotNil(t, version)

	assert.Equal(t, "", version)

	mockDockerSvc.AssertNumberOfCalls(t, "GetDockerTagsWithDate", 1)
}

func TestVerifyNewVersion_8(t *testing.T) {
	appContext := AppContext{}

	appContext.ChartImageCache.Store("repo/my-chart - 0.1.0", "myrepo.com/my-chart")

	latest := "20.2.1-RC-1"
	productVersion := "20.2.1-RC-0"
	svcCurrentVersion := "20.1.0-0"
	hotFix := false

	mockDockerSvc := mockGetDockerTagsWithDate(&appContext, getTagResponse(latest))
	version, err := appContext.verifyNewVersion("repo/my-chart - 0.1.0", svcCurrentVersion, productVersion, hotFix)
	assert.NoError(t, err)
	assert.NotNil(t, version)

	assert.Equal(t, latest, version)

	mockDockerSvc.AssertNumberOfCalls(t, "GetDockerTagsWithDate", 1)
}

func TestVerifyNewVersionHotfix_9(t *testing.T) {
	appContext := AppContext{}

	appContext.ChartImageCache.Store("repo/my-chart - 0.1.0", "myrepo.com/my-chart")

	latest := "20.1.1-10"
	productVersion := "20.1.1-1.1"
	svcCurrentVersion := "20.1.1-6"
	hotFix := true

	mockDockerSvc := mockGetDockerTagsWithDate(&appContext, getTagResponse(latest))
	version, err := appContext.verifyNewVersion("repo/my-chart - 0.1.0", svcCurrentVersion, productVersion, hotFix)
	assert.NoError(t, err)
	assert.NotNil(t, version)

	assert.Equal(t, "", version)

	mockDockerSvc.AssertNumberOfCalls(t, "GetDockerTagsWithDate", 1)
}

func TestVerifyNewVersionHotfix_10(t *testing.T) {
	appContext := AppContext{}

	appContext.ChartImageCache.Store("repo/my-chart - 0.1.0", "myrepo.com/my-chart")

	latest := "20.1.1-6.1"
	productVersion := "20.1.1-1.1"
	svcCurrentVersion := "20.1.1-6"
	hotFix := true

	mockDockerSvc := mockGetDockerTagsWithDate(&appContext, getTagResponse(latest))
	version, err := appContext.verifyNewVersion("repo/my-chart - 0.1.0", svcCurrentVersion, productVersion, hotFix)
	assert.NoError(t, err)
	assert.NotNil(t, version)

	assert.Equal(t, latest, version)

	mockDockerSvc.AssertNumberOfCalls(t, "GetDockerTagsWithDate", 1)
}

func TestGetMajorVersion(t *testing.T) {
	appContext := AppContext{}

	v1 := appContext.getMajorVersion("20.1.1-0")
	assert.Equal(t, "20.1.1", v1)

	v2 := appContext.getMajorVersion("20.1.1-0.1")
	assert.Equal(t, "20.1.1-0", v2)

	v3 := appContext.getMajorVersion("20.1.1-RC-0")
	assert.Equal(t, "20.1.1-RC", v3)
}

func TestGetMajorVersionOfHotfix1(t *testing.T) {
	appContext := AppContext{}
	v1 := appContext.getMajorVersionOfHotfix("20.1.1-0")
	assert.Equal(t, "20.1.1-0", v1)
}

func TestGetMajorVersionOfHotfix2(t *testing.T) {
	appContext := AppContext{}
	v2 := appContext.getMajorVersionOfHotfix("20.1.1-0.1")
	assert.Equal(t, "20.1.1-0", v2)
}

func TestGetMajorVersionOfHotfix3(t *testing.T) {
	appContext := AppContext{}
	v2 := appContext.getMajorVersionOfHotfix("20.1.1")
	assert.Equal(t, "20.1.1", v2)
}

func TestGetMinorVersion(t *testing.T) {
	appContext := AppContext{}

	v1 := appContext.getMinorVersion("20.1.1-0")
	assert.Equal(t, 0, v1)

	v2 := appContext.getMinorVersion("20.1.1-0.10")
	assert.Equal(t, 10, v2)

	v3 := appContext.getMinorVersion("20.1.1-RC-01")
	assert.Equal(t, 01, v3)
}

func TestGetMinorVersionOfHotfix1(t *testing.T) {
	appContext := AppContext{}
	v1 := appContext.getMinorVersionOfHotfix("20.1.1-0")
	assert.Equal(t, -1, v1)
}

func TestGetMinorVersionOfHotfix2(t *testing.T) {
	appContext := AppContext{}
	v2 := appContext.getMinorVersionOfHotfix("20.1.1-0.10")
	assert.Equal(t, 10, v2)
}

func TestGetMinorVersionOfHotfix3(t *testing.T) {
	appContext := AppContext{}
	v2 := appContext.getMinorVersionOfHotfix("20.1.1")
	assert.Equal(t, -1, v2)
}

func TestDiff(t *testing.T) {
	appContext := AppContext{}

	v1 := appContext.isDifferent(true, true, true)
	assert.Equal(t, false, v1)

	v2 := appContext.isDifferent(false, false, false)
	assert.Equal(t, false, v2)

	v3 := appContext.isDifferent(true, true, false)
	assert.Equal(t, true, v3)

	v4 := appContext.isDifferent(true, false, true)
	assert.Equal(t, true, v4)

	v5 := appContext.isDifferent(false, true, true)
	assert.Equal(t, true, v5)

	v6 := appContext.isDifferent(false, false, true)
	assert.Equal(t, true, v6)
}

func TestVerifyNewVersion_NotOk_1(t *testing.T) {
	appContext := AppContext{}

	appContext.ChartImageCache.Store("repo/my-chart - 0.1.0", "myrepo.com/my-chart")

	mockDockerSvc := mockGetDockerTagsWithDate(&appContext, getTagResponse("20.1.1-0"))

	version, err := appContext.verifyNewVersion("repo/my-chart - 0.1.0", "20.1.0-0", "20.1.0-0", false)
	assert.NoError(t, err)
	assert.NotNil(t, version)

	assert.Equal(t, "", version)

	mockDockerSvc.AssertNumberOfCalls(t, "GetDockerTagsWithDate", 1)
}

func TestVerifyNewVersion_NotOk_2(t *testing.T) {
	appContext := AppContext{}

	appContext.ChartImageCache.Store("foo", "bar")

	mockDockerSvc := mockGetDockerTagsWithDate(&appContext, getTagResponse("19.0.2-0"))
	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	appContext.HelmServiceAPI = mockHelmSvc

	bytes := []byte("{\"image\":{\"repository\":\"myrepo.com/my-chart\"}}")
	mockHelmSvc.On("GetValues", "repo/my-chart - 0.1.0", "0").Return(bytes, nil)

	version, err := appContext.verifyNewVersion("repo/my-chart - 0.1.0", "19.0.1-0", "19.0.1-0", false)
	assert.NoError(t, err)
	assert.NotNil(t, version)

	assert.Equal(t, "", version)

	mockDockerSvc.AssertNumberOfCalls(t, "GetDockerTagsWithDate", 1)
}

func TestVerifyNewVersion_NotOk_Error(t *testing.T) {
	appContext := AppContext{}

	appContext.ChartImageCache.Store("repo/my-chart - 0.1.0", "")

	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	appContext.HelmServiceAPI = mockHelmSvc

	mockHelmSvc.On("GetValues", mock.Anything, mock.Anything).Return(nil, errors.New("some error"))

	version, err := appContext.verifyNewVersion("repo/my-chart - 0.1.0", "19.0.1-0", "19.0.1-0", false)
	assert.Error(t, err)
	assert.NotNil(t, version)

	assert.Equal(t, "", version)
}

func TestVerifyNewVersion_NotOk_UnmarshalError(t *testing.T) {
	appContext := AppContext{}

	appContext.ChartImageCache.Store("repo/my-chart - 0.1.0", "")

	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	appContext.HelmServiceAPI = mockHelmSvc

	bytes := []byte(`["foo":"baz"]`)
	mockHelmSvc.On("GetValues", "repo/my-chart - 0.1.0", "0").Return(bytes, nil)

	version, err := appContext.verifyNewVersion("repo/my-chart - 0.1.0", "19.0.1-0", "19.0.1-0", false)
	assert.Error(t, err)
	assert.NotNil(t, version)
	assert.Equal(t, "", version)
}

func TestVerifyNewVersion_GetDockerTagsWithDateError(t *testing.T) {
	appContext := AppContext{}

	appContext.ChartImageCache.Store("repo/my-chart - 0.1.0", "myrepo.com/my-chart")

	mockDockerSvc := &mocks.DockerServiceInterface{}
	mockDockerSvc.On("GetDockerTagsWithDate", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("some error"))
	appContext.DockerServiceAPI = mockDockerSvc

	version, err := appContext.verifyNewVersion("repo/my-chart - 0.1.0", "19.0.1-0", "19.0.1-0", false)
	assert.Error(t, err)
	assert.NotNil(t, version)
	assert.Equal(t, "", version)

	mockDockerSvc.AssertNumberOfCalls(t, "GetDockerTagsWithDate", 1)
}

func TestVerifyNewVersion_NoNewVersion(t *testing.T) {
	appContext := AppContext{}

	appContext.ChartImageCache.Store("repo/my-chart - 0.1.0", "myrepo.com/my-chart")

	mockDockerSvc := mockGetDockerTagsWithDate(&appContext, getTagResponse("19.0.1-0"))

	version, err := appContext.verifyNewVersion("repo/my-chart - 0.1.0", "19.0.1-0", "19.0.1-0", false)
	assert.NoError(t, err)
	assert.NotNil(t, version)

	assert.Equal(t, "", version)

	mockDockerSvc.AssertNumberOfCalls(t, "GetDockerTagsWithDate", 1)
}

func TestValidateVersion_Valid1(t *testing.T) {
	appContext := AppContext{}
	valid := appContext.validateVersion("19.3.1-0", "19.3.1-1")
	assert.True(t, valid)
}

func TestValidateVersion_Valid2(t *testing.T) {
	appContext := AppContext{}
	valid := appContext.validateVersion("19.3.1-0", "19.3.1")
	assert.True(t, valid)
}

func TestValidateVersion_Valid3(t *testing.T) {
	appContext := AppContext{}
	valid := appContext.validateVersion("19.3.1", "19.3.1-1")
	assert.True(t, valid)
}

func TestValidateVersion_Valid4(t *testing.T) {
	appContext := AppContext{}
	valid := appContext.validateVersion("19.3.1", "19.3.1")
	assert.True(t, valid)
}

func TestValidateVersion_Valid5(t *testing.T) {
	appContext := AppContext{}
	valid := appContext.validateVersion("19.3.1-0", "19.3.1-0")
	assert.True(t, valid)
}

func TestValidateVersion_Invalid1(t *testing.T) {
	appContext := AppContext{}
	valid := appContext.validateVersion("19.3.1-0", "20.3.1-1")
	assert.False(t, valid)
}

func TestValidateVersion_Invalid2(t *testing.T) {
	appContext := AppContext{}
	valid := appContext.validateVersion("19.3.1-0", "19.4.1")
	assert.False(t, valid)
}

func TestValidateVersion_Invalid3(t *testing.T) {
	appContext := AppContext{}
	valid := appContext.validateVersion("19.3.1", "19.3.5-1")
	assert.False(t, valid)
}

func TestValidateVersion_Invalid4(t *testing.T) {
	appContext := AppContext{}
	valid := appContext.validateVersion("19.3.1", "19.3.11")
	assert.False(t, valid)
}

func TestValidateVersion_Invalid5(t *testing.T) {
	appContext := AppContext{}
	valid := appContext.validateVersion("19.3.1-0", "19.33.1-0")
	assert.False(t, valid)
}

func TestValidateVersion_Invalid6(t *testing.T) {
	appContext := AppContext{}
	valid := appContext.validateVersion("19.3.1-0", "19.3.11-0")
	assert.False(t, valid)
}

func TestValidateVersion_Invalid7(t *testing.T) {
	appContext := AppContext{}
	valid := appContext.validateVersion("19.3.1-0", "19.33.11-0")
	assert.False(t, valid)
}

func TestValidateVersion_Invalid8(t *testing.T) {
	appContext := AppContext{}
	valid := appContext.validateVersion("19.3.1.0", "1.33.11.0")
	assert.False(t, valid)
}

func TestValidateVersion_Invalid9(t *testing.T) {
	appContext := AppContext{}
	valid := appContext.validateVersion("19.31", "19")
	assert.False(t, valid)
}

func mockGetDockerTagsWithDate(appContext *AppContext, result *model.ListDockerTagsResult) *mocks.DockerServiceInterface {
	mockDockerSvc := &mocks.DockerServiceInterface{}
	mockDockerSvc.On("GetDockerTagsWithDate", mock.Anything, mock.Anything, mock.Anything).Return(result, nil)
	appContext.DockerServiceAPI = mockDockerSvc

	return mockDockerSvc
}

func getTagResponse(tag string) *model.ListDockerTagsResult {
	result := &model.ListDockerTagsResult{}

	var tr model.TagResponse
	tr.Tag = tag
	tr.Created = time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)
	result.TagResponse = append(result.TagResponse, tr)

	return result
}

func TestTriggerNewReleaseWebhookErrorListWebHooksByEnvAndType(t *testing.T) {
	appContext := AppContext{}

	mockWebHookDAO := &mockRepo.WebHookDAOInterface{}
	mockWebHookDAO.On("ListWebHooksByEnvAndType", -1, "HOOK_NEW_RELEASE").
		Return(nil, errors.New("error"))
	appContext.Repositories.WebHookDAO = mockWebHookDAO
	appContext.triggerNewReleaseWebhook(1, "release", 1)
	mockWebHookDAO.AssertNumberOfCalls(t, "ListWebHooksByEnvAndType", 1)
}

func TestTriggerNewReleaseWebhookErrorFindProductByID(t *testing.T) {
	appContext := AppContext{}

	webHooks := make([]model.WebHook, 0)
	webHooks = append(webHooks, mockWebHook())
	mockWebHookDAO := &mockRepo.WebHookDAOInterface{}
	mockWebHookDAO.On("ListWebHooksByEnvAndType", -1, "HOOK_NEW_RELEASE").
		Return(webHooks, nil)
	appContext.Repositories.WebHookDAO = mockWebHookDAO

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("FindProductByID", 999).Return(model.Product{}, errors.New("error"))
	appContext.Repositories.ProductDAO = mockProductDAO

	appContext.triggerNewReleaseWebhook(999, "release", 999)

	mockWebHookDAO.AssertNumberOfCalls(t, "ListWebHooksByEnvAndType", 1)
}
