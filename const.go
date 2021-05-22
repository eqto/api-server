package api

import "github.com/valyala/fasthttp"

const (
	//MethodGet GET
	MethodGet string = `GET`
	//MethodPost POST
	MethodPost string = `POST`

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
