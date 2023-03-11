package handler

import (
	"encoding/json"
	"family-catering/internal/model"
	"family-catering/internal/service"
	log "family-catering/pkg/logger"
	"family-catering/pkg/web"
	"fmt"
	"net/http"
)

type OrderHandler interface {
	Create() http.HandlerFunc
	ConfirmPayment() http.HandlerFunc
	Search() http.HandlerFunc
}

type orderHandler struct {
	orderService service.OrderService
}

func NewOrderHandler(orderService service.OrderService) OrderHandler {
	return &orderHandler{orderService: orderService}
}

// CreateOrder godoc
//	@Router			/order [post]
//	@Summary		Create order
//	@Description	Create a new order
//	@Tags			order
//	@Accept			json
//	@produce		json
//	@Param			Authorization	header		string																		true	"Insert your access token"	default(Bearer <your access token here>)
//	@param			payload			body		model.CreateOrderRequest													true	"body request"
//	@Success		200				{object}	web.JSONResponse{data=model.OrderResponse{order=model.CreateOrderResponse}}	"Ok"
//	@Failure		400				{object}	web.ErrJSONResponse															"Bad request"
//	@Failure		401				{object}	web.ErrJSONResponse															"Unauthorized"
//	@Failure		404				{object}	web.ErrJSONResponse															"Not found"
//	@Failure		422				{object}	web.ErrJSONResponse															"Unprocessable entity"
//	@Failure		500				{object}	web.ErrJSONResponse															"Internal server error"
func (handler *orderHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := web.RequestStartTimeFromContext(r.Context())
		req := model.CreateOrderRequest{}

		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			err := fmt.Errorf("handler.orderHandler.Create: %w", err)
			log.Error(err, "error unmarshal request")
			web.WriteFailJSON(w, http.StatusBadRequest, "error unmarshal request", start)
			return
		}

		resp, err := handler.orderService.Create(r.Context(), req)
		if err != nil {
			err = fmt.Errorf("handler.orderHandler.Create: %w", err)
			web.WriteHTTPError(w, err, start)
			return
		}

		payload := model.OrderResponse{Order: resp}
		web.WriteSuccessJSON(w, payload, start)

	}
}

// ConfirmPaymentOrder godoc
//	@Router			/order/confirm/ [put]
//	@Summary		Confirm order payment
//	@Description	Confirm unpaid order based on given email
//	@Tags			order
//	@produce		json
//	@Param			Authorization	header		string				true	"Insert your access token"	default(Bearer <your access token here>)
//	@param			payload			body		model.ConfirmPaymentRequest				true	"customer email"
//	@Success		200				{object}	web.JSONResponse	"Ok"
//	@Failure		400				{object}	web.ErrJSONResponse	"Bad request"
//	@Failure		401				{object}	web.ErrJSONResponse	"Unauthorized"
//	@Failure		404				{object}	web.ErrJSONResponse	"Not found"
//	@Failure		500				{object}	web.ErrJSONResponse	"Internal server error"
func (handler *orderHandler) ConfirmPayment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := web.RequestStartTimeFromContext(r.Context())
		// email := web.PathParamString(r, "email")
		req := model.ConfirmPaymentRequest{}
		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			err := fmt.Errorf("handler.orderHandler.Create: %w", err)
			log.Error(err, "error unmarshal request")
			web.WriteFailJSON(w, http.StatusBadRequest, "error unmarshal request", start)
			return
		}

		err = handler.orderService.ConfirmPayment(r.Context(), req)
		if err != nil {
			err = fmt.Errorf("handler.orderHandler.ConfirmPayment: %w", err)
			web.WriteHTTPError(w, err, start)
			return
		}
		web.WriteSuccessJSON(w, nil, start)
	}
}

// func (handler *orderHandler) CancelUnpaidOrder() http.HandlerFunc only be used for cronjob

// SearchOrder godoc
//	@Router			/order/search [get]
//	@Summary		Search order
//	@Description	Search an order based on given query params
//	@Tags			order
//	@Accept			json
//	@produce		json
//	@Param			Authorization	header		string				true	"Insert your access token"	default(Bearer <your access token here>)
//	@param			menu-names		query		string				false	"menu'ss name (if more than one name separated with comma without additional space)"
//	@param			emails			query		string				false	"customer's email (if more than one email separated with comma without additional space)"
//	@param			exact-names		query		string				false	"menu's name"
//	@param			qty				query		string				false	"the number of menus ordered per menu"
//	@param			max-price		query		string				false	"maximum price of menu"
//	@param			min-price		query		string				false	"minimum price of menu"
//	@param			status			query		string				false	"status or ordered menu"
//	@param			start-day		query		string				false	"ordered menu start at given day"
//	@param			end-day			query		string				false	"ordered menu end at given day"
//	@Success		200				{object}	web.JSONResponse	"Ok"
//	@Failure		400				{object}	web.ErrJSONResponse	"Bad request"
//	@Failure		401				{object}	web.ErrJSONResponse	"Unauthorized"
//	@Failure		404				{object}	web.ErrJSONResponse	"Not found"
//	@Failure		500				{object}	web.ErrJSONResponse	"Internal server error"
func (handler *orderHandler) Search() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := web.RequestStartTimeFromContext(r.Context())

		req, err := web.ParseSearchQueryParams(r)
		if err != nil {
			err := fmt.Errorf("handler.orderHandler.Search: %w", err)
			log.Error(err, "parse query params")
			web.WriteFailJSON(w, http.StatusBadRequest, "error parse query params", start)
			return
		}

		resp, err := handler.orderService.Search(r.Context(), req)
		if err != nil {
			err = fmt.Errorf("handler.orderHandler.Search: %w", err)
			web.WriteHTTPError(w, err, start)
			return
		}

		payload := model.OrderResponse{Order: resp}
		web.WriteSuccessJSON(w, payload, start)
	}
}
