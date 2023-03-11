package handler

import (
	"encoding/json"
	"family-catering/config"
	"family-catering/internal/model"
	"family-catering/internal/service"
	"family-catering/pkg/apperrors"
	"family-catering/pkg/consts"
	log "family-catering/pkg/logger"
	"family-catering/pkg/utils"
	"family-catering/pkg/web"
	"fmt"
	"net/http"
	"time"
)

type AuthHandler interface {
	Login() http.HandlerFunc
	Logout() http.HandlerFunc
	ForgotPassword() http.HandlerFunc
	RenewAccessToken() http.HandlerFunc

	AuthorizationRequired(next http.Handler) http.Handler
	SessionRequired(next http.Handler) http.Handler
}

type authHandler struct {
	authService service.AuthService
}

func NewAuthandler(authService service.AuthService) AuthHandler {
	return &authHandler{authService: authService}
}

// LoginAuth godoc
//	@Router			/auth/login [post]
//	@Summary		Login owner
//	@Description	Login owner using registered email and password
//	@Tags			auth
//	@Accept			json
//	@produce		json
//	@param			payload	body		model.AuthLoginRequest													true	"Create owner payload"
//	@Success		200		{object}	web.JSONResponse{data=model.AuthResponse{auth=model.AuthLoginResponse}}	"Ok"
//	@Failure		400		{object}	web.ErrJSONResponse														"Bad request"
//	@Failure		400		{object}	web.ErrJSONResponse{}													"Not Found"
//	@Failure		422		{object}	web.ErrJSONResponse														"Unprocessable entity"
//	@Failure		500		{object}	web.ErrJSONResponse														"Internal server error"
func (handler *authHandler) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := web.RequestStartTimeFromContext(r.Context())
		req := model.AuthLoginRequest{}

		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Error(err, "error unmarshal request's payload")
			web.WriteFailJSON(w, http.StatusBadRequest, "error unmarshal request's payload", start)
			return
		}

		resp, err := handler.authService.Login(r.Context(), req)

		if err != nil || resp == nil {
			web.WriteHTTPError(w, err, start)
			return
		}
		// set sid to cookie
		http.SetCookie(w, &http.Cookie{
			Name:     consts.CookieSID,
			Value:    resp.SID,
			HttpOnly: true,
			Expires:  time.Now().Add(15 * time.Minute), //? should use config?
		})

		payload := model.AuthResponse{Auth: resp}
		web.WriteSuccessJSON(w, payload, start)

	}
}

// LogoutAuth godoc
//	@Router			/auth/logout [delete]
//	@Summary		Logout owner
//	@Description	Logout owner
//	@Tags			auth
//	@Accept			json
//	@produce		json
//	@param			payload	body		model.AuthLogoutRequest	true	"logout owner payload"
//	@Success		200		{object}	web.JSONResponse{}		"Ok"
//	@Failure		400		{object}	web.ErrJSONResponse		"Bad request"
//	@Failure		401		{object}	web.ErrJSONResponse		"Unauthorized"
//	@Failure		422		{object}	web.ErrJSONResponse		"Unprocessable entity"
//	@Failure		500		{object}	web.ErrJSONResponse		"Internal server error"
func (handler *authHandler) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := web.RequestStartTimeFromContext(r.Context())
		req := model.AuthLogoutRequest{}

		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Error(err, "error unmarshal request's payload")
			web.WriteFailJSON(w, http.StatusBadRequest, "error unmarshal request payload", start)
			return
		}
		err = handler.authService.Logout(r.Context(), req)

		if err != nil {
			web.WriteHTTPError(w, err, start)
			return
		}

		web.WriteSuccessJSON(w, nil, start)
	}
}

// ForgotPasswordAuth godoc
//	@Router			/auth/forgot-password [put]
//	@Summary		Auth Forgot password
//	@Description	Auth Forgot password
//	@Tags			auth
//	@Accept			json
//	@produce		json
//	@param			payload	body		model.AuthForgotPasswordRequest	true	"logout owner payload"
//	@Success		200		{object}	web.JSONResponse				"Ok"
//	@Failure		400		{object}	web.ErrJSONResponse				"Bad request"
//	@Failure		401		{object}	web.ErrJSONResponse				"Unauthorized"
//	@Failure		404		{object}	web.ErrJSONResponse				"Not found"
//	@Failure		422		{object}	web.ErrJSONResponse				"Unprocessable entity"
//	@Failure		500		{object}	web.ErrJSONResponse				"Internal server error"
func (handler *authHandler) ForgotPassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := web.RequestStartTimeFromContext(r.Context())
		req := model.AuthForgotPasswordRequest{}

		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Error(err, "error unmarshal request's payload")
			web.WriteFailJSON(w, http.StatusBadRequest, "error unmarshal request payload", start)
			return
		}

		token, err := handler.authService.ForgotPassword(r.Context(), req)
		if err != nil {
			web.WriteHTTPError(w, err, start)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     consts.CookieResetPasswordToken, //  used to validate given access token
			Value:    token,
			HttpOnly: true,
			Expires:  time.Now().Add(config.Cfg().Web.AccessTokenTTL),
		})

		web.WriteSuccessJSON(w, nil, start)
	}
}

// RenewAccessTokenAuth godoc
//	@Router			/auth/renew-access-token [get]
//	@Summary		Auth Forgot password
//	@Description	Auth Forgot password
//	@Tags			auth
//	@Accept			json
//	@produce		json
//	@Success		200	{object}	web.JSONResponse{data=model.AuthResponse{auth=model.AuthRenewAccessTokenResponse}}	"Ok"
//	@Failure		401	{object}	web.ErrJSONResponse																	"Unauthorized"
//	@Failure		500	{object}	web.ErrJSONResponse																	"Internal server error"
func (handler *authHandler) RenewAccessToken() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := web.RequestStartTimeFromContext(r.Context())
		resp, err := handler.authService.RenewAccessToken(r.Context())
		if err != nil {
			web.WriteHTTPError(w, err, start)
			return
		}

		payload := model.AuthResponse{Auth: resp}
		web.WriteSuccessJSON(w, payload, start)

	}
}

func (handler *authHandler) AuthorizationRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := web.RequestStartTimeFromContext(r.Context())
		token := web.Authorization(r)
		if token == "" {
			err := fmt.Errorf("handler.auth.AuthorizationRequired: missing auth token")
			err = apperrors.WrapError(err, apperrors.ErrFieldValidationRequired, "missing auth token")
			log.Error(err, "missing token authorization")
			web.WriteFailJSON(w, http.StatusBadRequest, "missing auth token", start)
			return
		}

		*r = *r.WithContext(utils.ContextWithValue(r.Context(), consts.CtxKeyAuthorization, token))
		next.ServeHTTP(w, r)
	})
}

func (handler *authHandler) SessionRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := web.RequestStartTimeFromContext(r.Context())
		cookie, err := r.Cookie(consts.CookieSID)
		if err != nil {
			err := fmt.Errorf("handler.auth.SessionRequired: missing session id")
			err = apperrors.WrapError(err, apperrors.ErrFieldValidationRequired, "missing session id")
			log.Error(err, "missing  session id")
			web.WriteFailJSON(w, http.StatusBadRequest, "missing session id", start)
			return
		}
		sid := cookie.Value
		*r = *r.WithContext(utils.ContextWithValue(r.Context(), consts.CtxKeySID, sid))
		session, err := handler.authService.Session(r.Context(), sid)
		if err == nil && session == nil {
			err := fmt.Errorf("handler.auth.SessionRequired: empty session")
			err = apperrors.WrapError(err, apperrors.ErrFieldValidationRequired, "empty session")
			log.Error(err, "empty session")
			web.WriteFailJSON(w, http.StatusBadRequest, "empty session", start)
			return
		}
		if err != nil {
			err = fmt.Errorf("handler.authHandler.SessionRequired: %w", err)
			err = apperrors.WrapError(err, apperrors.ErrFieldValidationRequired, fmt.Sprintf("error get session, sid %s", sid))
			log.Error(err, "error get session")
			web.WriteFailJSON(w, http.StatusInternalServerError, "error get session", start)
			return
		}
		*r = *r.WithContext(utils.ContextWithValue(r.Context(), consts.CtxKeySession, session))
		next.ServeHTTP(w, r)
	})
}
