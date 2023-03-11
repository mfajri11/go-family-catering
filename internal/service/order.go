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

// const (
// new = iota + 1
// paid
// cancelled
// )

type OrderService interface {
	Create(ctx context.Context, req model.CreateOrderRequest) (resp *model.CreateOrderResponse, err error)
	Search(ctx context.Context, req model.OrderQuery) (resp *model.SearchOrdersResponse, err error)
	CancelUnpaidOrder(ctx context.Context) (resp *model.CancelUnpaidOrderResponse, err error)
	ConfirmPayment(ctx context.Context, req model.ConfirmPaymentRequest) error
}

type orderService struct {
	orderRepo repository.OrderRepository
	menuRepo  repository.MenuRepository
}

func NewOrderService(orderRepo repository.OrderRepository, menuRepo repository.MenuRepository) OrderService {
	return &orderService{orderRepo: orderRepo, menuRepo: menuRepo}
}

func (svc *orderService) Create(ctx context.Context, req model.CreateOrderRequest) (resp *model.CreateOrderResponse, err error) {
	token, ok := utils.ValueContext(ctx, consts.CtxKeyAuthorization).(string)
	if !ok {
		err := fmt.Errorf("service.orderService.Create: invalid auth token type want string got %T", token)
		return nil, apperrors.WrapError(err, apperrors.ErrAuth, "invalid auth token type")
	}
	_, err = utils.ValidateToken(token)
	if !errors.Is(err, nil) {
		err = fmt.Errorf("service.orderService.Create: %w", err)
		return nil, apperrors.WrapError(err, apperrors.ErrAuth, "")
	}
	err = utils.ValidateRequest(&req)
	if err == apperrors.ErrRequiredParam {
		return nil, apperrors.WrapError(err, apperrors.ErrFieldValidationRequired, "")
	}
	if err != nil {
		return nil, apperrors.WrapError(err, apperrors.ErrFieldValidation, "")
	}

	menusName := make([]string, 0, len(req.Orders))
	qtys := make(map[string]int, len(req.Orders)) // not sure order of rows from search menu, so this would be better for now
	for _, order := range req.Orders {
		menusName = append(menusName, order.Name)
		qtys[order.Name] = order.Qty
	}

	menus, errNoRow, err := svc.menuRepo.Search(ctx, model.MenuQuery{
		Names:           menusName,
		ExactNamesMatch: true,
	})

	if errNoRow != nil || len(menus) != len(req.Orders) {
		err = fmt.Errorf("service.orderService.Create: %w", err)
		return nil, apperrors.WrapError(err, apperrors.ErrNotFound, "")
	}

	if err != nil {
		err = fmt.Errorf("service.orderService.Create: %w", err)
		return nil, err
	}

	ordersDB := []*model.Order{}
	var totalPrice float32
	for _, menu := range menus {
		ordersDB = append(ordersDB, &model.Order{
			CustomerEmail: req.CustomerEmail,
			MenuName:      menu.Name,
			MenuID:        menu.ID,
			Price:         menu.Price,
			Qty:           qtys[menu.Name],
			Status:        consts.StatusNew,
		})

		totalPrice += (menu.Price * float32(qtys[menu.Name]))
	}

	_, orderID, err := svc.orderRepo.Create(ctx, ordersDB)

	if err != nil {
		err = fmt.Errorf("service.orderService.Create: %w", err)
		return nil, err
	}

	resp = &model.CreateOrderResponse{
		OrderID:       orderID,
		CustomerEmail: req.CustomerEmail,
		Message:       "success create orders",
		TotalPrice:    totalPrice,
	}

	return resp, nil
}

func (svc *orderService) Search(ctx context.Context, req model.OrderQuery) (resp *model.SearchOrdersResponse, err error) {
	token, ok := utils.ValueContext(ctx, consts.CtxKeyAuthorization).(string)
	if !ok {
		err := fmt.Errorf("service.orderService.Search: invalid auth token type want string got %T", token)
		return nil, apperrors.WrapError(err, apperrors.ErrAuth, "invalid auth token type")
	}
	_, err = utils.ValidateToken(token)
	if !errors.Is(err, nil) {
		err = fmt.Errorf("service.orderService.Search: %w", err)
		return nil, apperrors.WrapError(err, apperrors.ErrAuth, "")
	}

	ordersDB, errNoRow, err := svc.orderRepo.Search(ctx, req)
	if errNoRow != nil {
		errNoRow = fmt.Errorf("service.orderService.Search: %w", errNoRow)
		return nil, apperrors.WrapError(errNoRow, apperrors.ErrNotFound, "")
	}
	if err != nil {
		err = fmt.Errorf("service.orderService.Search: %w", err)
		return nil, err
	}

	searchedOrders := make([]*model.SearchResponse, 0, len(ordersDB))
	totalPrice := float32(0)
	for _, order := range ordersDB {
		searchedOrders = append(searchedOrders, &model.SearchResponse{
			OrderID:       order.OrderID,
			CustomerEmail: order.CustomerEmail,
			MenuName:      order.MenuName,
			MenuId:        order.MenuID,
			Price:         order.Price,
			Qty:           order.Qty,
			Status:        order.Status,
			CreatedAt:     order.CreatedAt,
		})

		totalPrice += float32(order.Qty) * order.Price
	}
	orders := model.SearchOrdersResponse{
		Orders:     searchedOrders,
		TotalPrice: totalPrice,
	}

	return &orders, nil
}

func (svc *orderService) CancelUnpaidOrder(ctx context.Context) (resp *model.CancelUnpaidOrderResponse, err error) {
	// will be used only by cron so no need to auth

	nAffected, err := svc.orderRepo.CancelUnpaidOrder(ctx)
	if err != nil {
		err = fmt.Errorf("service.orderService.CancelUnpaidOrder: %w", err)
		return nil, err
	}

	resp = &model.CancelUnpaidOrderResponse{
		Message:             "success cancel unpaid order",
		TotalOrderCancelled: nAffected,
	}

	return resp, nil
}

func (svc *orderService) ConfirmPayment(ctx context.Context, req model.ConfirmPaymentRequest) error {
	token, ok := utils.ValueContext(ctx, consts.CtxKeyAuthorization).(string)
	if !ok {
		err := fmt.Errorf("service.orderService.ConfirmPayment: invalid auth token type want string got %T", token)
		return apperrors.WrapError(err, apperrors.ErrAuth, "invalid auth token type")
	}
	_, err := utils.ValidateToken(token)
	if !errors.Is(err, nil) {
		err = fmt.Errorf("service.orderService.ConfirmPayment: %w", err)
		return apperrors.WrapError(err, apperrors.ErrAuth, "")
	}
	err = utils.ValidateRequest(&req)
	if err == apperrors.ErrRequiredParam {
		return apperrors.WrapError(err, apperrors.ErrFieldValidationRequired, "")
	}
	if err != nil {
		return apperrors.WrapError(err, apperrors.ErrFieldValidation, "")
	}

	_, errNoRow, err := svc.orderRepo.ConfirmPayment(ctx, req.Email)
	if errNoRow != nil && err == nil {
		errNoRow = fmt.Errorf("service.orderService.ConfirmPayment: %w", errNoRow)
		return apperrors.WrapError(errNoRow, apperrors.ErrNotFound, "")
	}

	if err != nil {
		err = fmt.Errorf("service.orderService.ConfirmPayment: %w", err)
		return err
	}

	return nil
}
