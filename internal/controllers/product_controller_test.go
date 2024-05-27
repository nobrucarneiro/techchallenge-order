package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/g73-techchallenge-order/internal/core/entities"
	"github.com/g73-techchallenge-order/internal/core/usecases/dto"
	mock_usecases "github.com/g73-techchallenge-order/internal/core/usecases/mocks"
	"github.com/g73-techchallenge-order/internal/infra/drivers/sql"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var productRequestMissingPrice, _ = os.ReadFile("./testdata/product_request_missing_price.json")
var productRequestValid, _ = os.ReadFile("./testdata/product_request_valid.json")
var productResponseValid, _ = os.ReadFile("./testdata/product_response_valid.json")

func TestProductController_GetProducts(t *testing.T) {
	ctrl := gomock.NewController(t)
	productUseCase := mock_usecases.NewMockProductUsecase(ctrl)
	productController := NewProductController(productUseCase)

	gin.SetMode(gin.TestMode)
	c, e := gin.CreateTestContext(httptest.NewRecorder())
	e.GET("/v1/products", productController.GetProducts)

	type args struct {
		category string
		limit    string
		offset   string
	}
	type want struct {
		statusCode int
		respBody   string
	}
	type productsUseCaseCall struct {
		category string
		times    int
		page     dto.Page[entities.Product]
		err      error
	}
	tests := []struct {
		name string
		args
		want
		productsUseCaseCall
	}{
		{
			name: "should return bad request when limit is not a number",
			args: args{
				limit:  "123abc",
				offset: "",
			},
			want: want{
				statusCode: 400,
				respBody:   `{"message":"invalid query parameters","error":"strconv.Atoi: parsing \"123abc\": invalid syntax"}`,
			},
		},
		{
			name: "should return bad request when offset is not a number",
			args: args{
				limit:  "1",
				offset: "123abc",
			},
			want: want{
				statusCode: 400,
				respBody:   `{"message":"invalid query parameters","error":"strconv.Atoi: parsing \"123abc\": invalid syntax"}`,
			},
		},
		{
			name: "should not get product by category when the user case returns error",
			args: args{
				category: "Acompanhamento",
				limit:    "1",
				offset:   "2",
			},
			want: want{
				statusCode: 500,
				respBody:   `{"message":"failed to get products by category","error":"internal server error"}`,
			},
			productsUseCaseCall: productsUseCaseCall{
				category: "Acompanhamento",
				times:    1,
				page:     dto.Page[entities.Product]{},
				err:      errors.New("internal server error"),
			},
		},
		{
			name: "should not get all products when the user case returns error",
			args: args{
				category: "",
				limit:    "1",
				offset:   "2",
			},
			want: want{
				statusCode: 500,
				respBody:   `{"message":"failed to get all products","error":"internal server error"}`,
			},
			productsUseCaseCall: productsUseCaseCall{
				category: "",
				times:    1,
				page:     dto.Page[entities.Product]{},
				err:      errors.New("internal server error"),
			},
		},
		{
			name: "should get products by category succesfully",
			args: args{
				category: "Acompanhamento",
				limit:    "1",
				offset:   "2",
			},
			want: want{
				statusCode: 200,
				respBody:   string(productResponseValid),
			},
			productsUseCaseCall: productsUseCaseCall{
				category: "Acompanhamento",
				times:    1,
				page: dto.Page[entities.Product]{
					Result: []entities.Product{
						{
							ID:          123,
							Name:        "Product 1",
							SkuId:       "33333",
							Description: "Description of product 1",
							Category:    "Acompanhamento",
							Price:       9.99,
							CreatedAt:   time.Time{},
							UpdatedAt:   time.Time{},
						},
					},
					Next: new(int),
				},
				err: nil,
			},
		},
		{
			name: "should get all products succesfully",
			args: args{
				limit:  "1",
				offset: "2",
			},
			want: want{
				statusCode: 200,
				respBody:   string(productResponseValid),
			},
			productsUseCaseCall: productsUseCaseCall{
				times: 1,
				page: dto.Page[entities.Product]{
					Result: []entities.Product{
						{
							ID:          123,
							Name:        "Product 1",
							SkuId:       "33333",
							Description: "Description of product 1",
							Category:    "Acompanhamento",
							Price:       9.99,
							CreatedAt:   time.Time{},
							UpdatedAt:   time.Time{},
						},
					},
					Next: new(int),
				},
				err: nil,
			},
		},
	}

	for _, tt := range tests {
		if tt.args.category != "" {
			productUseCase.
				EXPECT().
				GetProductsByCategory(gomock.Any(), gomock.Eq(tt.productsUseCaseCall.category)).
				Times(tt.productsUseCaseCall.times).
				Return(tt.productsUseCaseCall.page, tt.productsUseCaseCall.err)
		} else {
			productUseCase.
				EXPECT().
				GetAllProducts(gomock.Any()).
				Times(tt.productsUseCaseCall.times).
				Return(tt.productsUseCaseCall.page, tt.productsUseCaseCall.err)
		}

		c.Request, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/v1/products?limit=%s&offset=%s&category=%s", tt.args.limit, tt.args.offset, tt.args.category), nil)
		c.Request.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		e.ServeHTTP(rr, c.Request)

		assert.Equal(t, tt.want.statusCode, rr.Code)
		assert.Equal(t, tt.want.respBody, rr.Body.String())
	}
}

func TestProductController_CreateProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	productUseCase := mock_usecases.NewMockProductUsecase(ctrl)
	productController := NewProductController(productUseCase)

	gin.SetMode(gin.TestMode)
	c, e := gin.CreateTestContext(httptest.NewRecorder())
	e.POST("/v1/products", productController.CreateProducts)

	type args struct {
		reqBody string
	}
	type want struct {
		statusCode int
		respBody   string
	}
	type productUseCaseCall struct {
		times int
		err   error
	}
	tests := []struct {
		name string
		args
		want
		productUseCaseCall
	}{
		{
			name: "should return bad request when req body is not a json",
			args: args{
				reqBody: "<invalidJson>",
			},
			want: want{
				statusCode: 400,
				respBody:   `{"message":"failed to bind product payload","error":"invalid character '\u003c' looking for beginning of value"}`,
			},
		},
		{
			name: "should return bad request when name is missing in the request",
			args: args{
				reqBody: string(productRequestMissingPrice),
			},
			want: want{
				statusCode: 400,
				respBody:   `{"message":"invalid product payload","error":"price: non zero value required"}`,
			},
		},
		{
			name: "should not create product when the user case returns error",
			args: args{
				reqBody: string(productRequestValid),
			},
			want: want{
				statusCode: 500,
				respBody:   `{"message":"failed to create product","error":"internal server error"}`,
			},
			productUseCaseCall: productUseCaseCall{
				times: 1,
				err:   errors.New("internal server error"),
			},
		},
		{
			name: "should create product succesfully",
			args: args{
				reqBody: string(productRequestValid),
			},
			want: want{
				statusCode: 200,
				respBody:   "",
			},
			productUseCaseCall: productUseCaseCall{
				times: 1,
				err:   nil,
			},
		},
	}

	for _, tt := range tests {
		productUseCase.
			EXPECT().
			CreateProduct(gomock.Any()).
			Times(tt.productUseCaseCall.times).
			Return(tt.productUseCaseCall.err)

		c.Request, _ = http.NewRequest(http.MethodPost, "/v1/products", strings.NewReader(tt.args.reqBody))
		c.Request.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		e.ServeHTTP(rr, c.Request)

		assert.Equal(t, tt.want.statusCode, rr.Code)
		assert.Equal(t, tt.want.respBody, rr.Body.String())
	}
}

func TestProductController_UpdateProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	productUseCase := mock_usecases.NewMockProductUsecase(ctrl)
	productController := NewProductController(productUseCase)

	gin.SetMode(gin.TestMode)
	c, e := gin.CreateTestContext(httptest.NewRecorder())
	e.PUT("/v1/products", productController.UpdateProduct)
	e.PUT("/v1/products/:id", productController.UpdateProduct)

	type args struct {
		id      string
		reqBody string
	}
	type want struct {
		statusCode int
		respBody   string
	}
	type productUseCaseCall struct {
		productId string
		times     int
		err       error
	}
	tests := []struct {
		name string
		args
		want
		productUseCaseCall
	}{
		{
			name: "should return bad request when id is missing",
			args: args{
				id:      "",
				reqBody: string(productRequestValid),
			},
			want: want{
				statusCode: 400,
				respBody:   `{"message":"id path param is required","error":"id path parameter is missing"}`,
			},
		},
		{
			name: "should return bad request when req body is not a json",
			args: args{
				id:      "222",
				reqBody: "<invalidJson>",
			},
			want: want{
				statusCode: 400,
				respBody:   `{"message":"failed to bind product payload","error":"invalid character '\u003c' looking for beginning of value"}`,
			},
		},
		{
			name: "should return bad request when product payload is missing price",
			args: args{
				id:      "222",
				reqBody: string(productRequestMissingPrice),
			},
			want: want{
				statusCode: 400,
				respBody:   `{"message":"invalid product payload","error":"price: non zero value required"}`,
			},
		},
		{
			name: "should not update product when the user case returns error",
			args: args{
				id:      "222",
				reqBody: string(productRequestValid),
			},
			want: want{
				statusCode: 500,
				respBody:   `{"message":"failed to update product","error":"internal server error"}`,
			},
			productUseCaseCall: productUseCaseCall{
				productId: "222",
				times:     1,
				err:       errors.New("internal server error"),
			},
		},
		{
			name: "should not update product when the product is not found",
			args: args{
				id:      "222",
				reqBody: string(productRequestValid),
			},
			want: want{
				statusCode: 404,
				respBody:   `{"message":"product not found","error":"entity not found"}`,
			},
			productUseCaseCall: productUseCaseCall{
				productId: "222",
				times:     1,
				err:       sql.ErrNotFound,
			},
		},
		{
			name: "should update product succesfully",
			args: args{
				id:      "222",
				reqBody: string(productRequestValid),
			},
			want: want{
				statusCode: 200,
				respBody:   "",
			},
			productUseCaseCall: productUseCaseCall{
				productId: "222",
				times:     1,
				err:       nil,
			},
		},
	}

	for _, tt := range tests {
		productUseCase.
			EXPECT().
			UpdateProduct(gomock.Eq(tt.productUseCaseCall.productId), gomock.Any()).
			Times(tt.productUseCaseCall.times).
			Return(tt.productUseCaseCall.err)

		pathParam := ""
		if tt.args.id != "" {
			pathParam = "/" + tt.args.id
		}
		c.Request, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/v1/products%s", pathParam), strings.NewReader(tt.reqBody))
		c.Request.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		e.ServeHTTP(rr, c.Request)

		assert.Equal(t, tt.want.statusCode, rr.Code)
		assert.Equal(t, tt.want.respBody, rr.Body.String())
	}
}

func TestProductController_DeleteProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	productUseCase := mock_usecases.NewMockProductUsecase(ctrl)
	productController := NewProductController(productUseCase)

	gin.SetMode(gin.TestMode)
	c, e := gin.CreateTestContext(httptest.NewRecorder())
	e.DELETE("/v1/products", productController.DeleteProduct)
	e.DELETE("/v1/products/:id", productController.DeleteProduct)

	type args struct {
		id string
	}
	type want struct {
		statusCode int
		respBody   string
	}
	type productUseCaseCall struct {
		productId string
		times     int
		err       error
	}
	tests := []struct {
		name string
		args
		want
		productUseCaseCall
	}{
		{
			name: "should return bad request when id is missing",
			args: args{
				id: "",
			},
			want: want{
				statusCode: 400,
				respBody:   `{"message":"id path param is required","error":"id path parameter is missing"}`,
			},
		},

		{
			name: "should not delete product when the user case returns error",
			args: args{
				id: "222",
			},
			want: want{
				statusCode: 500,
				respBody:   `{"message":"failed to delete product","error":"internal server error"}`,
			},
			productUseCaseCall: productUseCaseCall{
				productId: "222",
				times:     1,
				err:       errors.New("internal server error"),
			},
		},
		{
			name: "should not delete product when the product is not found",
			args: args{
				id: "222",
			},
			want: want{
				statusCode: 404,
				respBody:   `{"message":"product not found","error":"entity not found"}`,
			},
			productUseCaseCall: productUseCaseCall{
				productId: "222",
				times:     1,
				err:       sql.ErrNotFound,
			},
		},
		{
			name: "should delete product succesfully",
			args: args{
				id: "222",
			},
			want: want{
				statusCode: 204,
				respBody:   "",
			},
			productUseCaseCall: productUseCaseCall{
				productId: "222",
				times:     1,
				err:       nil,
			},
		},
	}

	for _, tt := range tests {
		productUseCase.
			EXPECT().
			DeleteProduct(gomock.Eq(tt.productUseCaseCall.productId)).
			Times(tt.productUseCaseCall.times).
			Return(tt.productUseCaseCall.err)

		pathParam := ""
		if tt.args.id != "" {
			pathParam = "/" + tt.args.id
		}
		c.Request, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/v1/products%s", pathParam), nil)
		c.Request.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		e.ServeHTTP(rr, c.Request)

		assert.Equal(t, tt.want.statusCode, rr.Code)
		assert.Equal(t, tt.want.respBody, rr.Body.String())
	}
}
