package api

import "github.com/valyala/fasthttp"

const (
	//MethodGet GET
	MethodGet string = fasthttp.MethodGet
	//MethodPost POST
	MethodPost string = fasthttp.MethodPost

	//StatusBadRequest ...
	StatusBadRequest = fasthttp.StatusBadRequest
	//StatusUnauthorized ...
	StatusUnauthorized = fasthttp.StatusUnauthorized
	//StatusNotFound ...
	StatusForbidden = fasthttp.StatusForbidden
	//StatusNotFound ...
	StatusNotFound = fasthttp.StatusNotFound
	//StatusOK ...
	StatusOK = 200

	//StatusInternalServerError ...
	StatusInternalServerError = fasthttp.StatusInternalServerError
	StatusServiceUnavailable  = fasthttp.StatusServiceUnavailable

	//StatusBadGateway ...
	StatusBadGateway = 502
)
