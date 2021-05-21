package api

import "github.com/valyala/fasthttp"

const (
	//MethodGet GET
	MethodGet string = `GET`
	//MethodPost POST
	MethodPost string = `POST`

	//StatusBadRequest ...
	StatusBadRequest = 400
	//StatusUnauthorized ...
	StatusUnauthorized = 401
	//StatusNotFound ...
	StatusNotFound = fasthttp.StatusNotFound
	//StatusOK ...
	StatusOK = 200

	//StatusInternalServerError ...
	StatusInternalServerError = 500
	StatusServiceUnavailable  = fasthttp.StatusServiceUnavailable

	//StatusBadGateway ...
	StatusBadGateway = 502
)
