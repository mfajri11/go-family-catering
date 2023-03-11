package handler

import (
	"encoding/json"
	"errors"
	"family-catering/internal/model"
	"family-catering/internal/service"
	log "family-catering/pkg/logger"
	"family-catering/pkg/web"
	"fmt"
	"net/http"
)

type MenuHandler interface {
	GetByID() http.HandlerFunc
	GetByName() http.HandlerFunc
	List() http.HandlerFunc
	Create() http.HandlerFunc
	Update() http.HandlerFunc
	Delete() http.HandlerFunc
}

type menuHandler struct {
	menuService service.MenuService
}

// authorization token assume exists on context passed by authHandler.Authorize middleware

func NewMenuHandler(menuService service.MenuService) MenuHandler {
	return &menuHandler{menuService: menuService}
}

// GetMenuByID godoc
//	@Router			/menu/{id} [get]
//	@Summary		Get menu
//	@Description	Show interest menu detail by given id
//	@Tags			menu
//	@Param			Authorization	header	string	true	"Insert your access token"	default(Bearer <your access token here>)
//	@param			id				path	int		true	"Menu id"					Format(int64)
//	@Produce		json
//	@Success		200	{object}	web.JSONResponse{data=model.MenuResponse{menu=model.GetMenuResponse}}	"Ok"
//	@Failure		500	{object}	web.ErrJSONResponse														"Internal server error"
//	@Failure		400	{object}	web.ErrJSONResponse														"Bad request"
//	@Failure		404	{object}	web.ErrJSONResponse														"Menu not found"
//	@Failure		401	{object}	web.ErrJSONResponse														"Unauthorized"
func (handler *menuHandler) GetByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := web.RequestStartTimeFromContext(r.Context())
		id, err := web.PathParamInt64(r, "id")
		if !errors.Is(err, nil) {
			err := fmt.Errorf("handler.menuHandler.GetByID: %w", err)
			log.Error(err, "invalid path params")
			web.WriteFailJSON(w, http.StatusBadRequest, "invalid path params", start)
			return
		}

		menu, err := handler.menuService.GetByID(r.Context(), id)
		if err != nil {
			web.WriteHTTPError(w, err, start)
			return
		}

		payload := model.MenuResponse{Menu: menu}
		web.WriteSuccessJSON(w, payload, start)

	}
}

// GetMenuByName godoc
//	@Router			/menu/{name} [get]
//	@Summary		Get menu by given name
//	@Description	Show interest menu detail by given name
//	@Tags			menu
//	@Param			Authorization	header	string	true	"Insert your access token"	default(Bearer <your access token here>)
//	@param			name			path	string	true	"Menu name"
//	@Produce		json
//	@Success		200	{object}	web.JSONResponse{data=model.MenuResponse{menu=model.GetMenuResponse}}	"Ok"
//	@Failure		500	{object}	web.ErrJSONResponse														"Internal server error"
//	@Failure		400	{object}	web.ErrJSONResponse														"Bad request"
//	@Failure		404	{object}	web.ErrJSONResponse														"Menu not found"
//	@Failure		401	{object}	web.ErrJSONResponse														"Unauthorized"
func (handler *menuHandler) GetByName() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		start := web.RequestStartTimeFromContext(r.Context())
		name := web.PathParamString(r, "name")
		if name == "" {
			err := fmt.Errorf("handler.menuHandler.GetByName: missing path params")
			log.Error(err, "invalid path params")
			web.WriteFailJSON(w, http.StatusBadRequest, "missing path params", start)
			return
		}

		menu, err := handler.menuService.GetByName(r.Context(), name)
		if err != nil {
			web.WriteHTTPError(w, err, start)
			return
		}

		payload := model.MenuResponse{Menu: menu}
		web.WriteSuccessJSON(w, payload, start)

	}

}

// ListMenu godoc
//	@Router			/menu [get]
//	@Summary		Show list of menus
//	@Description	Show list of menus by (optionally) by given limit of offset
//	@Tags			menu
//	@Param			Authorization	header	string	true	"Insert your access token"	default(Bearer <your access token here>)
//	@param			limit			query	int		false	"Pagination limit"			Format(int64)
//	@param			offset			query	int		false	"Pagination offset"			Format(int64)
//	@Produce		json
//	@Success		200	{object}	web.JSONResponse{data=model.MenuResponse{menu=[]model.GetMenuResponse}}	"Ok"
//	@Failure		500	{object}	web.ErrJSONResponse														"Internal server error"
//	@Failure		400	{object}	web.ErrJSONResponse														"Bad request"
func (handler *menuHandler) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := web.RequestStartTimeFromContext(r.Context())
		limit, offset, err := web.PaginationLimitOffset(r)
		if !errors.Is(err, nil) {
			err := fmt.Errorf("handler.menuHandler.List: %w", err)
			log.Error(err, "invalid query params")
			web.WriteFailJSON(w, http.StatusBadRequest, "invalid query params", start)
			return
		}

		menus, err := handler.menuService.List(r.Context(), limit, offset)
		if err != nil {
			web.WriteHTTPError(w, err, start)
			return
		}

		payload := model.MenuResponse{Menu: menus}
		web.WriteSuccessJSON(w, payload, start)
	}
}

// CreateMenu godoc
//	@Router			/menu [post]
//	@Summary		Create a menu
//	@Description	Create a new menu
//	@Tags			menu
//	@Accept			json
//	@produce		json
//	@Param			Authorization	header		string																		true	"Insert your access token"	default(Bearer <your access token here>)
//	@param			payload			body		model.CreateMenuRequest														true	"body request"
//	@Success		200				{object}	web.JSONResponse{data=model.MenuResponse{menu=model.CreateMenuResponse}}	"Ok"
//	@Failure		500				{object}	web.ErrJSONResponse															"Internal server error"
//	@Failure		400				{object}	web.ErrJSONResponse															"Bad request"
//	@Failure		422				{object}	web.ErrJSONResponse															"Unprocessable entity"
func (handler *menuHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := web.RequestStartTimeFromContext(r.Context())
		req := model.CreateMenuRequest{}

		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			err := fmt.Errorf("handler.menuHandler.Create: %w", err)
			log.Error(err, "error unmarshal request")
			web.WriteFailJSON(w, http.StatusBadRequest, "error unmarshal request", start)
			return
		}

		menu, err := handler.menuService.Create(r.Context(), req)
		if err != nil {
			web.WriteHTTPError(w, err, start)
			return
		}

		payload := model.MenuResponse{Menu: menu}
		web.WriteSuccessJSON(w, payload, start)
	}
}

// CreateMenu godoc
//	@Router			/menu/{id} [put]
//	@Summary		Update  menu
//	@Description	Update menu by given id
//	@Tags			menu
//	@Accept			json
//	@produce		json
//	@param			id				path		int																			true	"Menu id"					Format(int64)
//	@Param			Authorization	header		string																		true	"Insert your access token"	default(Bearer <your access token here>)
//	@param			payload			body		model.UpdateMenuRequest														true	"body request"
//	@Success		200				{object}	web.JSONResponse{data=model.MenuResponse{menu=model.UpdateMenuResponse}}	"Ok"
//	@Failure		400				{object}	web.ErrJSONResponse															"Bad request"
//	@Failure		401				{object}	web.ErrJSONResponse															"Unauthorized"
//	@Failure		422				{object}	web.ErrJSONResponse															"Unprocessable entity"
//	@Failure		500				{object}	web.ErrJSONResponse															"Internal server error"
func (handler *menuHandler) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := web.RequestStartTimeFromContext(r.Context())
		req := model.UpdateMenuRequest{}

		id, err := web.PathParamInt64(r, "id")
		if !errors.Is(err, nil) {
			err := fmt.Errorf("handler.menuHandler.Update: %w", err)
			log.Error(err, "invalid path params")
			web.WriteFailJSON(w, http.StatusBadRequest, "invalid path params", start)
			return
		}
		defer r.Body.Close()
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			err := fmt.Errorf("handler.menuHandler.Update: %w", err)
			log.Error(err, "error unmarshal request")
			web.WriteFailJSON(w, http.StatusBadRequest, "error unmarshal request", start)
			return
		}

		menu, err := handler.menuService.Update(r.Context(), id, req)
		if err != nil {
			web.WriteHTTPError(w, err, start)
			return
		}
		payload := model.MenuResponse{Menu: menu}
		web.WriteSuccessJSON(w, payload, start)

	}
}

// DeleteOwner godoc
//	@Router			/menu/{id} [delete]
//	@Summary		Delete menu
//	@Description	Delete menu by given id
//	@Tags			menu
//	@param			id				path	int		true	"Menu id"					Format(int64)
//	@Param			Authorization	header	string	true	"Insert your access token"	default(Bearer <your access token here>)
//	@Produce		json
//	@Success		200	{object}	web.JSONResponse	required	"Ok"
//	@Failure		500	{object}	web.ErrJSONResponse	"Internal server error"
//	@Failure		400	{object}	web.ErrJSONResponse	"Bad request"
//	@Failure		401	{object}	web.ErrJSONResponse	"Unauthorized"
//	@Failure		404	{object}	web.ErrJSONResponse	"Menu not found"
func (handler *menuHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := web.RequestStartTimeFromContext(r.Context())

		id, err := web.PathParamInt64(r, "id")
		if !errors.Is(err, nil) {
			err := fmt.Errorf("handler.menuHandler.Update: %w", err)
			log.Error(err, "invalid path params")
			web.WriteFailJSON(w, http.StatusBadRequest, "invalid path params", start)
			return
		}

		nAffected, err := handler.menuService.Delete(r.Context(), id)
		if err != nil && nAffected <= 0 {
			web.WriteHTTPError(w, err, start)
			return
		}

		web.WriteSuccessJSON(w, nil, start)
	}
}
