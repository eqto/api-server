package api

import "github.com/valyala/fasthttp"

const (
	MethodGet  string = fasthttp.MethodGet
	MethodPost string = fasthttp.MethodPost

	StatusBadRequest   = fasthttp.StatusBadRequest
	StatusUnauthorized = fasthttp.StatusUnauthorized
	StatusForbidden    = fasthttp.StatusForbidden
	StatusNotFound     = fasthttp.StatusNotFound
	StatusOK           = fasthttp.StatusOK

	StatusInternalServerError = fasthttp.StatusInternalServerError
	StatusServiceUnavailable  = fasthttp.StatusServiceUnavailable

	StatusBadGateway = fasthttp.StatusBadGateway
)
