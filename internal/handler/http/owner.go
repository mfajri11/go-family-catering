package handler

import (
	"encoding/json"
	"errors"
	"family-catering/internal/model"
	"family-catering/internal/service"
	"family-catering/pkg/web"
	"fmt"
	"net/http"

	log "family-catering/pkg/logger"
)

type OwnerHandler interface {
	Create() http.HandlerFunc
	Get() http.HandlerFunc
	List() http.HandlerFunc
	Delete() http.HandlerFunc
	Update() http.HandlerFunc
	ResetPasswordById() http.HandlerFunc
	ResetPasswordByEmail() http.HandlerFunc
	UpdateEmailByID() http.HandlerFunc
}

type ownerHandler struct {
	ownerService service.OwnerService
}

func NewOwnerHandler(ownerService service.OwnerService) OwnerHandler {
	return &ownerHandler{ownerService: ownerService}
}

// CreateOwner godoc
//	@Router			/owner [post]
//	@Summary		Create an owner
//	@Description	Create an owner through Sign up
//	@Tags			owner
//	@Accept			json
//	@produce		json
//	@param			payload	body		model.CreateOwnerRequest													true	"Create owner payload"
//	@Success		200		{object}	web.JSONResponse{data=model.OwnerResponse{owner=model.CreateOwnerResponse}}	"Ok"
//	@Failure		400		{object}	web.ErrJSONResponse															"Bad request"
//	@Failure		409		{object}	web.ErrJSONResponse															"Email already registered"
//	@Failure		422		{object}	web.ErrJSONResponse															"Unprocessable entity"
//	@Failure		500		{object}	web.ErrJSONResponse															"Internal server error"
func (handler *ownerHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := web.RequestStartTimeFromContext(r.Context())
		req := model.CreateOwnerRequest{}

		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Error(err, "error unmarshal request")
			web.WriteFailJSON(w, http.StatusBadRequest, "error unmarshal request", start)
			return
		}

		owner, err := handler.ownerService.Create(r.Context(), req)

		if err != nil {
			web.WriteHTTPError(w, err, start)
			return
		}

		payload := model.OwnerResponse{Owner: owner}
		web.WriteSuccessJSON(w, payload, start)

	}
}

// GetOwner godoc
//	@Router			/owner/{id} [get]
//	@Summary		Get owner by given id
//	@Description	Show interest owner details
//	@Tags			owner
//	@param			id	path	int	true	"Owner id"	Format(int64)
//	@Produce		json
//	@Success		200	{object}	web.JSONResponse{data=model.OwnerResponse{owner=model.GetOwnerResponse}}	"Ok"
//	@Failure		400	{object}	web.ErrJSONResponse															"Bad request"
//	@Failure		401	{object}	web.ErrJSONResponse															"Unauthorized"
//	@Failure		404	{object}	web.ErrJSONResponse															"Owner not found"
//	@Failure		500	{object}	web.ErrJSONResponse															"Internal server error"
func (handler *ownerHandler) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := web.RequestStartTimeFromContext(r.Context())

		id, err := web.PathParamInt64(r, "id")
		if !errors.Is(err, nil) {
			err = fmt.Errorf("handler.ownerHandler.Get: %w", err)
			log.Error(err, "invalid path params")
			web.WriteFailJSON(w, http.StatusBadRequest, "invalid path params", start)
			return
		}

		owner, err := handler.ownerService.Get(r.Context(), id)
		if err != nil {
			err := fmt.Errorf("handler.ownerHandler.Get: %w", err)
			web.WriteHTTPError(w, err, start)
			return
		}

		payload := model.OwnerResponse{Owner: owner}
		web.WriteSuccessJSON(w, payload, start)
	}
}

// ListOwner godoc
//	@Router			/owner [get]
//	@Summary		Show list of owners
//	@Description	Show list of owners by (optionally) given limit and/or offset
//	@Tags			owner
//	@param			limit	query	int	false	"Pagination limit"	Format(int64)
//	@param			offset	query	int	false	"Pagination offset"	Format(int64)
//	@Produce		json
//	@Success		200	{object}	web.JSONResponse{data=model.OwnerResponse{owner=[]model.GetOwnerResponse}}	"Ok"
//	@Failure		400	{object}	web.ErrJSONResponse															"Bad request"
//	@Failure		500	{object}	web.ErrJSONResponse															"Internal server error"
func (handler *ownerHandler) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := web.RequestStartTimeFromContext(r.Context())

		limit, offset, err := web.PaginationLimitOffset(r)
		if !errors.Is(err, nil) {
			err := fmt.Errorf("handler.ownerHandler.List: %w", err)
			log.Error(err, "invalid query params")
			web.WriteFailJSON(w, http.StatusBadRequest, "invalid query params", start)
			return
		}

		owners, err := handler.ownerService.List(r.Context(), limit, offset)
		if err != nil {
			err := fmt.Errorf("handler.ownerHandler.List: %w", err)
			web.WriteHTTPError(w, err, start)
		}

		payload := model.OwnerResponse{Owner: owners}
		web.WriteSuccessJSON(w, payload, start)

	}
}

// DeleteOwner godoc
//	@Router			/owner/{id} [delete]
//	@Summary		Delete owner
//	@Description	Delete owner by given owner's id
//	@Tags			owner
//	@param			id				path	int		true	"Owner id"					Format(int64)
//	@Param			Authorization	header	string	true	"Insert your access token"	default(Bearer <Add access token here>)
//	@param			Cookie			header	string	true	"Session id"				default(sid=<Add your session id>)
//	@Produce		json
//	@Success		200	{object}	web.JSONResponse	required	"Ok"
//	@Failure		400	{object}	web.ErrJSONResponse	"Bad request"
//	@Failure		404	{object}	web.ErrJSONResponse	"Owner not found"
//	@Failure		500	{object}	web.ErrJSONResponse	"Internal server error"
func (handler *ownerHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := web.RequestStartTimeFromContext(r.Context())

		id, err := web.PathParamInt64(r, "id")
		if !errors.Is(err, nil) {
			err := fmt.Errorf("handler.ownerHandler.Delete: %w", err)
			log.Error(err, "invalid path params")
			web.WriteFailJSON(w, http.StatusBadRequest, "invalid path params", start)
			return
		}

		nAffected, err := handler.ownerService.Delete(r.Context(), id)

		if err != nil && nAffected <= 0 {
			err := fmt.Errorf("handler.ownerHandler.Delete: %w", err)
			web.WriteHTTPError(w, err, start)
			return
		}
		web.WriteSuccessJSON(w, nil, start)

	}
}

// UpdateOwner godoc
//	@Router			/owner/{id} [put]
//	@Summary		Update owner
//	@Description	Update owner by given owner's id
//	@Tags			owner
//	@param			id				path	int							true	"Owner id"					Format(int64)
//	@Param			Authorization	header	string						true	"Insert your access token"	default(Bearer <Add access token here>)
//	@param			payload			body	model.UpdateOwnerRequest	true	"Update owner payload"
//	@Produce		json
//	@Success		200	{object}	web.JSONResponse	required	"Ok"
//	@Failure		400	{object}	web.ErrJSONResponse	"Bad request"
//	@Failure		404	{object}	web.ErrJSONResponse	"Owner not found"
//	@Failure		422	{object}	web.ErrJSONResponse	"Unprocessable entity"
//	@Failure		500	{object}	web.ErrJSONResponse	"Internal server error"
func (handler *ownerHandler) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := web.RequestStartTimeFromContext(r.Context())
		req := model.UpdateOwnerRequest{}

		id, err := web.PathParamInt64(r, "id")
		if !errors.Is(err, nil) {
			err := fmt.Errorf("handler.ownerHandler.Update: %w", err)
			log.Error(err, "invalid path params")
			web.WriteFailJSON(w, http.StatusBadRequest, "invalid path params", start)
			return
		}

		defer r.Body.Close()
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			err := fmt.Errorf("handler.ownerHandler.Update: %w", err)
			log.Error(err, "error unmarshal request's payload")
			web.WriteFailJSON(w, http.StatusBadRequest, "error unmarshal request payload", start)
			return
		}

		owner, err := handler.ownerService.Update(r.Context(), id, req)
		if err != nil {
			err := fmt.Errorf("handler.ownerHandler.Update: %w", err)
			web.WriteHTTPError(w, err, start)
			return
		}

		payload := model.OwnerResponse{Owner: owner}
		web.WriteSuccessJSON(w, payload, start)
	}

}

// ResetPasswordByEmailOwner godoc
//	@Router			/owner/reset-password/{rpid} [put]
//	@Summary		Reset owner password (logged out state)
//	@Description	Reset owner password by given rpid
//	@Tags			owner
//	@param			rpid			path	string						true	"Owner reset password id"
//	@Param			Authorization	header	string						true	"Insert your password token"	default(Bearer <Add password token here>)
//	@param			payload			body	model.ResetPasswordRequest	true	"Reset Password payload"
//	@Produce		json
//	@Success		200	{object}	web.JSONResponse{}	required	"Ok"
//	@Failure		400	{object}	web.ErrJSONResponse	"Bad request"
//	@Failure		401	{object}	web.ErrJSONResponse	"Unauthorized"
//	@Failure		404	{object}	web.ErrJSONResponse	"Owner not found"
//	@Failure		422	{object}	web.ErrJSONResponse	"Unprocessable entity"
//	@Failure		500	{object}	web.ErrJSONResponse	"Internal server error"
func (handler *ownerHandler) ResetPasswordByEmail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := web.RequestStartTimeFromContext(r.Context())
		req := model.ResetPasswordRequest{}

		passwordResetId := web.PathParamString(r, "rpid")
		// rpid = path parameter for identifier request reset password
		// rpt reset password token used as cookie name which stored password-token

		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			err := fmt.Errorf("handler.ownerHandler.ResetPasswordByEmail: %w", err)
			log.Error(err, "error unmarshal request's payload")
			web.WriteFailJSON(w, http.StatusBadRequest, "error unmarshal request payload", start)
			return
		}

		err = handler.ownerService.ResetPasswordByEmail(r.Context(), passwordResetId, req)
		if err != nil {
			err := fmt.Errorf("handler.ownerHandler.ResetPasswordByEmail: %w", err)
			web.WriteHTTPError(w, err, start)
			return
		}

		web.WriteSuccessJSON(w, nil, start)
	}
}

// ResetPasswordByIDOwner godoc
//	@Router			/owner/{id}/reset-password [put]
//	@Summary		Reset owner password (logged in state)
//	@Description	Reset owner password by given id
//	@Tags			owner
//	@param			id				path	int							true	"Owner reset password id"	Format(int64)
//	@Param			Authorization	header	string						true	"Insert your access token"	default(Bearer <Add access token here>)
//	@param			Cookie			header	string						true	"Session id"				default(sid=<Add your session id>)
//	@param			payload			body	model.ResetPasswordRequest	true	"Reset Password payload"
//	@Produce		json
//	@Success		200	{object}	web.JSONResponse	required	"Ok"
//	@Failure		400	{object}	web.ErrJSONResponse	"Bad request"
//	@Failure		401	{object}	web.ErrJSONResponse	"Unauthorized"
//	@Failure		404	{object}	web.ErrJSONResponse	"Owner not found"
//	@Failure		422	{object}	web.ErrJSONResponse	"Unprocessable entity"
//	@Failure		500	{object}	web.ErrJSONResponse	"Internal server error"
func (handler *ownerHandler) ResetPasswordById() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := web.RequestStartTimeFromContext(r.Context())
		req := model.ResetPasswordRequest{}

		id, err := web.PathParamInt64(r, "id")
		if !errors.Is(err, nil) {
			err := fmt.Errorf("handler.ownerHandler.ResetPasswordByID: %w", err)
			log.Error(err, "invalid path params")
			web.WriteFailJSON(w, http.StatusBadRequest, "invalid path params", start)
			return
		}

		defer r.Body.Close()
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			err := fmt.Errorf("handler.ownerHandler.ResetPasswordByID: %w", err)
			log.Error(err, "error unmarshal request's payload")
			web.WriteFailJSON(w, http.StatusBadRequest, "error unmarshal request payload", start)
			return
		}

		err = handler.ownerService.ResetPasswordByID(r.Context(), id, req)
		if err != nil {
			web.WriteHTTPError(w, err, start)
			return
		}

		web.WriteSuccessJSON(w, nil, start)
	}
}

// UpdateEmailByIDOwner godoc
//	@Router			/owner/{id}/update-email [put]
//	@Summary		Update Email owner
//	@Description	Update owner's email by given id
//	@Tags			owner
//	@param			id				path	int		true	"Owner reset password id"	Format(int64)
//	@Param			Authorization	header	string	true	"Insert your access token"	default(Bearer <Add access token here>)
//	@param			Cookie			header	string	true	"Session id"				default(sid=<Add your session id>)
//	@Produce		json
//	@Success		200	{object}	web.JSONResponse	required	"Ok"
//	@Failure		400	{object}	web.ErrJSONResponse	"Bad request"
//	@Failure		401	{object}	web.ErrJSONResponse	"Unauthorized"
//	@Failure		404	{object}	web.ErrJSONResponse	"Owner not found"
//	@Failure		422	{object}	web.ErrJSONResponse	"Unprocessable entity"
//	@Failure		500	{object}	web.ErrJSONResponse	"Internal server error"
func (handler *ownerHandler) UpdateEmailByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := web.RequestStartTimeFromContext(r.Context())
		req := model.UpdateEmailRequest{}

		id, err := web.PathParamInt64(r, "id")
		if !errors.Is(err, nil) {
			err := fmt.Errorf("handler.ownerHandler.UpdateEmailByID: %w", err)
			log.Error(err, "invalid path params")
			web.WriteFailJSON(w, http.StatusBadRequest, "invalid path params", start)
			return
		}

		defer r.Body.Close()
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			err := fmt.Errorf("handler.ownerHandler.UpdateEmailByID: %w", err)
			log.Error(err, "error unmarshal request's payload")
			web.WriteFailJSON(w, http.StatusBadRequest, "error unmarshal request payload", start)
			return
		}

		err = handler.ownerService.UpdateEmailByID(r.Context(), id, req)
		if err != nil {
			web.WriteHTTPError(w, err, start)
			return
		}

		web.WriteSuccessJSON(w, nil, start)
	}
}
