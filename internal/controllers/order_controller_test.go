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
	"github.com/g73-techchallenge-order/internal/infra/drivers/authorizer"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var orderRequestMissingStatus, _ = os.ReadFile("./testdata/order_request_missing_status.json")
var orderRequestWrongCpf, _ = os.ReadFile("./testdata/order_request_wrong_cpf.json")
var orderRequestValid, _ = os.ReadFile("./testdata/order_request_valid.json")
var orderResponseValid, _ = os.ReadFile("./testdata/order_response_valid.json")

func TestOrderController_CreateOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	orderUseCase := mock_usecases.NewMockOrderUsecase(ctrl)
	orderController := NewOrderController(orderUseCase)

	gin.SetMode(gin.TestMode)
	c, e := gin.CreateTestContext(httptest.NewRecorder())
	e.POST("/v1/orders", orderController.CreateOrder)

	type args struct {
		reqBody string
	}
	type want struct {
		statusCode int
		respBody   string
	}
	type orderUseCaseCall struct {
		times         int
		orderResponse dto.OrderCreationResponse
		err           error
	}
	tests := []struct {
		name string
		args
		want
		orderUseCaseCall
	}{
		{
			name: "should return bad request when req body is not a json",
			args: args{
				reqBody: "<invalidJson>",
			},
			want: want{
				statusCode: 400,
				respBody:   `{"message":"failed to bind order payload","error":"invalid character '\u003c' looking for beginning of value"}`,
			},
		},
		{
			name: "should return bad request when status is missing in the request",
			args: args{
				reqBody: string(orderRequestMissingStatus),
			},
			want: want{
				statusCode: 400,
				respBody:   `{"message":"invalid order payload","error":"Status is invalid"}`,
			},
		},
		{
			name: "should return bad request when cpf is wrong in the request",
			args: args{
				reqBody: string(orderRequestWrongCpf),
			},
			want: want{
				statusCode: 400,
				respBody:   `{"message":"invalid order payload","error":"invalid CPF [11122233344]"}`,
			},
		},
		{
			name: "should not authorize request when the user is not authorized",
			args: args{
				reqBody: string(orderRequestValid),
			},
			want: want{
				statusCode: 403,
				respBody:   `{"message":"customer cpf invalid","error":"customer unauthorized"}`,
			},
			orderUseCaseCall: orderUseCaseCall{
				times:         1,
				orderResponse: dto.OrderCreationResponse{},
				err:           authorizer.ErrUnauthorized,
			},
		},
		{
			name: "should not create order when the user case returns error",
			args: args{
				reqBody: string(orderRequestValid),
			},
			want: want{
				statusCode: 500,
				respBody:   `{"message":"failed to create order","error":"internal server error"}`,
			},
			orderUseCaseCall: orderUseCaseCall{
				times:         1,
				orderResponse: dto.OrderCreationResponse{},
				err:           errors.New("internal server error"),
			},
		},
		{
			name: "should create order succesfully",
			args: args{
				reqBody: string(orderRequestValid),
			},
			want: want{
				statusCode: 200,
				respBody:   `{"qrCode":"mercadopago123456","orderId":98765}`,
			},
			orderUseCaseCall: orderUseCaseCall{
				times: 1,
				orderResponse: dto.OrderCreationResponse{
					QRCode:  "mercadopago123456",
					OrderID: 98765,
				},
				err: nil,
			},
		},
	}

	for _, tt := range tests {
		orderUseCase.
			EXPECT().
			CreateOrder(gomock.Any()).
			Times(tt.orderUseCaseCall.times).
			Return(tt.orderUseCaseCall.orderResponse, tt.orderUseCaseCall.err)

		c.Request, _ = http.NewRequest(http.MethodPost, "/v1/orders", strings.NewReader(tt.args.reqBody))
		c.Request.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		e.ServeHTTP(rr, c.Request)

		assert.Equal(t, tt.want.statusCode, rr.Code)
		assert.Equal(t, tt.want.respBody, rr.Body.String())
	}
}

func TestOrderController_GetAllOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	orderUseCase := mock_usecases.NewMockOrderUsecase(ctrl)
	orderController := NewOrderController(orderUseCase)

	gin.SetMode(gin.TestMode)
	c, e := gin.CreateTestContext(httptest.NewRecorder())
	e.GET("/v1/orders", orderController.GetAllOrders)

	type args struct {
		limit  string
		offset string
	}
	type want struct {
		statusCode int
		respBody   string
	}
	type orderUseCaseCall struct {
		times int
		page  dto.Page[entities.Order]
		err   error
	}
	tests := []struct {
		name string
		args
		want
		orderUseCaseCall
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
			name: "should not get order when the user case returns error",
			args: args{
				limit:  "1",
				offset: "2",
			},
			want: want{
				statusCode: 500,
				respBody:   `{"message":"failed to get all orders","error":"internal server error"}`,
			},
			orderUseCaseCall: orderUseCaseCall{
				times: 1,
				page:  dto.Page[entities.Order]{},
				err:   errors.New("internal server error"),
			},
		},
		{
			name: "should get all orders succesfully",
			args: args{
				limit:  "1",
				offset: "2",
			},
			want: want{
				statusCode: 200,
				respBody:   string(orderResponseValid),
			},
			orderUseCaseCall: orderUseCaseCall{
				times: 1,
				page: dto.Page[entities.Order]{
					Result: []entities.Order{createOrder()},
					Next:   new(int),
				},
				err: nil,
			},
		},
	}

	for _, tt := range tests {
		orderUseCase.
			EXPECT().
			GetAllOrders(gomock.Any()).
			Times(tt.orderUseCaseCall.times).
			Return(tt.orderUseCaseCall.page, tt.orderUseCaseCall.err)

		c.Request, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/v1/orders?limit=%s&offset=%s", tt.args.limit, tt.args.offset), nil)
		c.Request.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		e.ServeHTTP(rr, c.Request)

		assert.Equal(t, tt.want.statusCode, rr.Code)
		assert.Equal(t, tt.want.respBody, rr.Body.String())
	}
}

func TestOrderController_GetOrderStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	orderUseCase := mock_usecases.NewMockOrderUsecase(ctrl)
	orderController := NewOrderController(orderUseCase)

	gin.SetMode(gin.TestMode)
	c, e := gin.CreateTestContext(httptest.NewRecorder())
	e.GET("/v1/orders/:id/status", orderController.GetOrderStatus)

	type args struct {
		id string
	}
	type want struct {
		statusCode int
		respBody   string
	}
	type orderUseCaseCall struct {
		orderId     int
		times       int
		orderStatus dto.OrderStatusDTO
		err         error
	}
	tests := []struct {
		name string
		args
		want
		orderUseCaseCall
	}{
		{
			name: "should return bad request when id is missing",
			args: args{},
			want: want{
				statusCode: 400,
				respBody:   `{"message":"[id] path parameter is required","error":"id is missing"}`,
			},
		},
		{
			name: "should return bad request when id is not a number",
			args: args{
				id: "abc",
			},
			want: want{
				statusCode: 400,
				respBody:   `{"message":"[id] path parameter is invalid","error":"strconv.Atoi: parsing \"abc\": invalid syntax"}`,
			},
		},
		{
			name: "should not create order when the user case returns error",
			args: args{
				id: "123",
			},
			want: want{
				statusCode: 500,
				respBody:   `{"message":"failed to get order status","error":"internal server error"}`,
			},
			orderUseCaseCall: orderUseCaseCall{
				orderId:     123,
				orderStatus: dto.OrderStatusDTO{},
				times:       1,
				err:         errors.New("internal server error"),
			},
		},
		{
			name: "should get order status succesfully",
			args: args{
				id: "123",
			},
			want: want{
				statusCode: 200,
				respBody:   `{"status":"CREATED"}`,
			},
			orderUseCaseCall: orderUseCaseCall{
				orderId: 123,
				orderStatus: dto.OrderStatusDTO{
					Status: "CREATED",
				},
				times: 1,
				err:   nil,
			},
		},
	}

	for _, tt := range tests {
		orderUseCase.
			EXPECT().
			GetOrderStatus(gomock.Eq(tt.orderUseCaseCall.orderId)).
			Times(tt.orderUseCaseCall.times).
			Return(tt.orderUseCaseCall.orderStatus, tt.orderUseCaseCall.err)

		c.Request, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/v1/orders/%s/status", tt.args.id), nil)
		c.Request.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		e.ServeHTTP(rr, c.Request)

		assert.Equal(t, tt.want.statusCode, rr.Code)
		assert.Equal(t, tt.want.respBody, rr.Body.String())
	}
}

func TestOrderController_UpdateOrderStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	orderUseCase := mock_usecases.NewMockOrderUsecase(ctrl)
	orderController := NewOrderController(orderUseCase)

	gin.SetMode(gin.TestMode)
	c, e := gin.CreateTestContext(httptest.NewRecorder())
	e.PUT("/v1/orders/:id/status", orderController.UpdateOrderStatus)

	type args struct {
		id      string
		reqBody string
	}
	type want struct {
		statusCode int
		respBody   string
	}
	type orderUseCaseCall struct {
		orderId     int
		orderStatus string
		times       int
		err         error
	}
	tests := []struct {
		name string
		args
		want
		orderUseCaseCall
	}{
		{
			name: "should return bad request when id is missing",
			args: args{
				reqBody: `{"status":"CREATED"}`,
			},
			want: want{
				statusCode: 400,
				respBody:   `{"message":"[id] path parameter is required","error":"id is missing"}`,
			},
		},
		{
			name: "should return bad request when id is not a number",
			args: args{
				id:      "abc",
				reqBody: `{"status":"CREATED"}`,
			},
			want: want{
				statusCode: 400,
				respBody:   `{"message":"[id] path parameter is invalid","error":"strconv.Atoi: parsing \"abc\": invalid syntax"}`,
			},
		},
		{
			name: "should return bad request when req body is not a json",
			args: args{
				id:      "123",
				reqBody: "<invalidJson>",
			},
			want: want{
				statusCode: 400,
				respBody:   `{"message":"failed to bind order status payload","error":"invalid character '\u003c' looking for beginning of value"}`,
			},
		},

		{
			name: "should return bad request when state is wrong in the request",
			args: args{
				id:      "123",
				reqBody: `{"status":"WRONG_STATE"}`,
			},
			want: want{
				statusCode: 400,
				respBody:   `{"message":"invalid order status payload","error":"status: WRONG_STATE does not validate as in(CREATED|PAID|RECEIVED|IN_PROGRESS|READY|DONE)"}`,
			},
		},
		{
			name: "should not create order when the user case returns error",
			args: args{
				id:      "123",
				reqBody: `{"status":"CREATED"}`,
			},
			want: want{
				statusCode: 500,
				respBody:   `{"message":"failed to update order status","error":"internal server error"}`,
			},
			orderUseCaseCall: orderUseCaseCall{
				orderId:     123,
				orderStatus: "CREATED",
				times:       1,
				err:         errors.New("internal server error"),
			},
		},
		{
			name: "should update order status succesfully",
			args: args{
				id:      "123",
				reqBody: string(orderRequestValid),
			},
			want: want{
				statusCode: 204,
				respBody:   "",
			},
			orderUseCaseCall: orderUseCaseCall{
				orderId:     123,
				orderStatus: "CREATED",
				times:       1,
				err:         nil,
			},
		},
	}

	for _, tt := range tests {
		orderUseCase.
			EXPECT().
			UpdateOrderStatus(gomock.Eq(tt.orderUseCaseCall.orderId), gomock.Eq(tt.orderUseCaseCall.orderStatus)).
			Times(tt.orderUseCaseCall.times).
			Return(tt.orderUseCaseCall.err)

		c.Request, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/v1/orders/%s/status", tt.args.id), strings.NewReader(tt.reqBody))
		c.Request.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		e.ServeHTTP(rr, c.Request)

		assert.Equal(t, tt.want.statusCode, rr.Code)
		assert.Equal(t, tt.want.respBody, rr.Body.String())
	}
}

func createOrder() entities.Order {
	return entities.Order{
		ID: 123,
		Items: []entities.OrderItem{
			{
				ID:       999,
				Quantity: 1,
				Type:     "UNIT",
				Product: entities.Product{
					ID:          222,
					Name:        "Batata Frita",
					SkuId:       "333",
					Description: "Batata canoa",
					Category:    "Acompanhamento",
					Price:       9.99,
					CreatedAt:   time.Time{},
					UpdatedAt:   time.Time{},
				},
			},
		},
		Coupon:      "APP10",
		TotalAmount: 9.99,
		Status:      "PAID",
		CreatedAt:   time.Time{},
		CustomerCPF: "111222333444",
	}
}
