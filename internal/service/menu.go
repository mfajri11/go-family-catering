package service

import (
	"context"
	"errors"
	"family-catering/internal/model"
	"family-catering/internal/repository"
	"family-catering/pkg/apperrors"
	"family-catering/pkg/consts"
	"family-catering/pkg/utils"
	"fmt"
)

type MenuService interface {
	GetByID(ctx context.Context, id int64) (*model.GetMenuResponse, error)
	GetByName(ctx context.Context, name string) (*model.GetMenuResponse, error)
	List(ctx context.Context, limit, offset int) ([]*model.GetMenuResponse, error)
	Create(ctx context.Context, req model.CreateMenuRequest) (*model.CreateMenuResponse, error)
	Update(ctx context.Context, id int64, req model.UpdateMenuRequest) (*model.UpdateMenuResponse, error)
	Delete(ctx context.Context, id int64) (nAffected int64, err error)
}

type menuService struct {
	menuRepo repository.MenuRepository
}

func NewMenuService(menuRepo repository.MenuRepository) MenuService {
	return &menuService{menuRepo: menuRepo}
}

func (svc *menuService) GetByID(ctx context.Context, id int64) (*model.GetMenuResponse, error) {
	// Authorization
	token, ok := utils.ValueContext(ctx, consts.CtxKeyAuthorization).(string)
	if !ok {
		err := fmt.Errorf("service.menuService.GetByID: invalid auth token type want string got %T", token)
		return nil, apperrors.WrapError(err, apperrors.ErrAuth, "invalid auth token type")
	}
	_, err := utils.ValidateToken(token)
	if !errors.Is(err, nil) {
		err := fmt.Errorf("service.menuService.GetByID: %w", err)
		return nil, apperrors.WrapError(err, apperrors.ErrAuth, "")
	}

	menu, errNoRow, err := svc.menuRepo.GetByID(ctx, id)
	if errNoRow != nil {
		errNoRow := fmt.Errorf("service.menuService.GetByID: %w", errNoRow)
		return nil, apperrors.WrapError(errNoRow, apperrors.ErrNotFound, "")
	}

	if err != nil {
		err := fmt.Errorf("service.menuService.GetByID: %w", err)
		return nil, err
	}

	return newMenuResponse(menu), nil
}

func (svc *menuService) GetByName(ctx context.Context, name string) (*model.GetMenuResponse, error) {
	// Authorization
	token, ok := utils.ValueContext(ctx, consts.CtxKeyAuthorization).(string)
	if !ok {
		err := fmt.Errorf("service.menuService.GetByName: invalid auth token type want string got %T", token)
		return nil, apperrors.WrapError(err, apperrors.ErrAuth, "invalid auth token type")
	}
	_, err := utils.ValidateToken(token)
	if !errors.Is(err, nil) {
		err := fmt.Errorf("service.menuService.GetByName: %w", err)
		return nil, apperrors.WrapError(err, apperrors.ErrAuth, "")
	}

	menu, errNoRow, err := svc.menuRepo.GetByName(ctx, name)
	if errNoRow != nil {
		errNoRow := fmt.Errorf("service.menuService.GetByName: %w", errNoRow)
		return nil, apperrors.WrapError(errNoRow, apperrors.ErrNotFound, "")
	}

	if err != nil {
		err := fmt.Errorf("service.menuService.GetByName: %w", err)
		return nil, err
	}

	return newMenuResponse(menu), nil
}

func (svc *menuService) List(ctx context.Context, limit, offset int) ([]*model.GetMenuResponse, error) {
	// Authorization
	token, ok := utils.ValueContext(ctx, consts.CtxKeyAuthorization).(string)
	if !ok {
		err := fmt.Errorf("service.menuService.List: invalid token type want string, got %T", token)
		return nil, apperrors.WrapError(err, apperrors.ErrAuth, "")
	}
	_, err := utils.ValidateToken(token)
	if !errors.Is(err, nil) {
		err := fmt.Errorf("service.menuService.List: %w", err)
		return nil, apperrors.WrapError(err, apperrors.ErrAuth, "")
	}

	menus, errNoRow, err := svc.menuRepo.List(ctx, limit, offset)

	if errNoRow != nil && err == nil {
		return []*model.GetMenuResponse{}, nil
	}

	if err != nil {
		err := fmt.Errorf("service.menuService.List: %w", err)
		return nil, err
	}

	return newMenusResponse(menus), err

}
func (svc *menuService) Create(ctx context.Context, req model.CreateMenuRequest) (*model.CreateMenuResponse, error) {
	// Authorization
	token, ok := utils.ValueContext(ctx, consts.CtxKeyAuthorization).(string)
	if !ok {
		err := fmt.Errorf("service.menuService.Create: invalid auth token type want string got %T", token)
		return nil, apperrors.WrapError(err, apperrors.ErrAuth, "invalid auth token type")
	}
	_, err := utils.ValidateToken(token)
	if !errors.Is(err, nil) {
		err := fmt.Errorf("service.menuService.Create: %w", err)
		return nil, apperrors.WrapError(err, apperrors.ErrAuth, "")
	}

	err = utils.ValidateRequest(&req)
	if errors.Is(err, apperrors.ErrRequiredParam) {
		err = fmt.Errorf("service.menuService.Create: %w", err)
		return nil, apperrors.WrapError(err, apperrors.ErrFieldValidationRequired, "")
	}
	if !errors.Is(err, nil) {
		err = fmt.Errorf("service.menuService.Create: %w", err)
		return nil, apperrors.WrapError(err, apperrors.ErrFieldValidation, "")
	}

	menu := model.Menu{
		Name:       req.Name,
		Price:      req.Price,
		Categories: req.Categories,
	}

	id, err := svc.menuRepo.Create(ctx, menu)
	if err != nil {
		err = fmt.Errorf("service.menuService.Create: %w", err)
		return nil, err

	}
	menu.ID = id
	return newMenuResponse(&menu), nil
}

func (svc *menuService) Update(ctx context.Context, id int64, req model.UpdateMenuRequest) (*model.UpdateMenuResponse, error) {
	// Authorization
	token, ok := utils.ValueContext(ctx, consts.CtxKeyAuthorization).(string)
	if !ok {
		err := fmt.Errorf("service.menuService.Update: invalid auth token type want string got %T", token)
		return nil, apperrors.WrapError(err, apperrors.ErrAuth, "invalid auth token type")
	}
	_, err := utils.ValidateToken(token)
	if !errors.Is(err, nil) {
		err := fmt.Errorf("service.menuService.Update: %w", err)
		return nil, apperrors.WrapError(err, apperrors.ErrAuth, "")
	}

	// request validation
	err = utils.ValidateRequest(&req)
	if errors.Is(err, apperrors.ErrRequiredParam) {
		err := fmt.Errorf("service.menuService.Update: %w", err)
		return nil, apperrors.WrapError(err, apperrors.ErrFieldValidationRequired, "")
	}
	if !errors.Is(err, nil) {
		err := fmt.Errorf("service.menuService.Update: %w", err)
		return nil, apperrors.WrapError(err, apperrors.ErrFieldValidation, "")
	}

	menu := model.Menu{
		ID:         id,
		Name:       req.Name,
		Price:      req.Price,
		Categories: req.Categories,
	}

	_, errNoRow, err := svc.menuRepo.Update(ctx, menu)
	if errNoRow != nil {
		err := fmt.Errorf("service.menuService.Update: %w", errNoRow)
		return nil, err
	}

	if err != nil {
		err := fmt.Errorf("service.menuService.Update: %w", err)
		return nil, err
	}

	return newMenuResponse(&menu), nil
}

func (svc *menuService) Delete(ctx context.Context, id int64) (nAffected int64, err error) {
	// Authorization
	token, ok := utils.ValueContext(ctx, consts.CtxKeyAuthorization).(string)
	if !ok {
		err := fmt.Errorf("service.menuService.Delete: invalid auth token type want string got %T", token)
		return 0, apperrors.WrapError(err, apperrors.ErrAuth, "invalid auth token type")
	}
	_, err = utils.ValidateToken(token)
	if !errors.Is(err, nil) {
		err = fmt.Errorf("service.menuService.Delete: %w", err)
		return 0, apperrors.WrapError(err, apperrors.ErrAuth, "")
	}

	nAffected, errNoRow, err := svc.menuRepo.Delete(ctx, id)
	if errNoRow != nil && nAffected <= 0 {
		err = fmt.Errorf("service.menuService.Delete: %w", err)
		return 0, apperrors.WrapError(err, apperrors.ErrNotFound, "")
	}
	if err != nil {
		err := fmt.Errorf("service.menuService.Delete: %w", err)
		return 0, err
	}

	return nAffected, nil
}
