package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/longjoy/micro-go-book/ch6-discovery/string-service/endpoint"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ErrorBadRequest = errors.New("invalid request parameter")
)

// MakeHttpHandler make http handler use mux
func MakeHttpHandler(ctx context.Context, endpoints endpoint.StringEndpoints, logger log.Logger) http.Handler {
	r := mux.NewRouter()

	options := []kithttp.ServerOption{
		kithttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		kithttp.ServerErrorEncoder(encodeError),
	}

	/*
		1. 通过decodeStringRequest将请求参数解析为endpoint能够识别的结构体
		2. 通过endpoint进行业务处理
		3. 其中endpoint中调用service层的方法，实现具体的业务逻辑
		4. 通过encodeStringResponse将响应结果编码为http.ResponseWriter，并返回

		总结：http请求 -> decodeStringRequest解析为endpoint能够识别的结构体 -> endpoint进行业务处理 -> service实际处理 -> 返回给endpoint -> encodeStringResponse编码为http.ResponseWriter并返回
	*/
	r.Methods("POST").Path("/op/{type}/{a}/{b}").Handler(kithttp.NewServer(
		endpoints.StringEndpoint,
		decodeStringRequest,
		encodeStringResponse,
		options...,
	))

	r.Path("/metrics").Handler(promhttp.Handler())

	// create health check handler
	r.Methods("GET").Path("/health").Handler(kithttp.NewServer(
		endpoints.HealthCheckEndpoint,
		decodeHealthCheckRequest,
		encodeStringResponse,
		options...,
	))

	return r
}

// decodeStringRequest decode request params to struct
func decodeStringRequest(_ context.Context, r *http.Request) (interface{}, error) {
	// 从请求中获取路由变量
	vars := mux.Vars(r)

	// 从路由变量中获取"type"字段
	requestType, ok := vars["type"]
	if !ok {
		// 如果未找到"type"字段，则返回错误
		return nil, ErrorBadRequest
	}

	// 从路由变量中获取"a"字段
	pa, ok := vars["a"]
	if !ok {
		// 如果未找到"a"字段，则返回错误
		return nil, ErrorBadRequest
	}

	// 从路由变量中获取"b"字段
	pb, ok := vars["b"]
	if !ok {
		// 如果未找到"b"字段，则返回错误
		return nil, ErrorBadRequest
	}

	// 返回包含请求类型、A和B的StringRequest结构体
	return endpoint.StringRequest{
		RequestType: requestType,
		A:           pa,
		B:           pb,
	}, nil
}

// encodeStringResponse encode response to return
func encodeStringResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

// decodeHealthCheckRequest decode request
func decodeHealthCheckRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return endpoint.HealthRequest{}, nil
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
