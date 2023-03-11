package repository

import (
	"context"
	"database/sql"
	"family-catering/internal/model"
	"family-catering/pkg/db/postgres"
	"fmt"
	"strings"
)

// const (
// 	orderStatusNEW = iota + 1
// 	orderStatusPAID
// 	orderStatusCANCELLED
// )

type OrderRepository interface {
	Search(ctx context.Context, order model.OrderQuery) (orders []*model.Order, errNoRow error, err error)
	Create(ctx context.Context, orders []*model.Order) (lastInsertbaseOrderID int64, OrderID int64, err error)
	ConfirmPayment(ctx context.Context, email string) (nAffected int64, errNoRow error, err error)
	// Report(ctx context.Context) // by id email, price and data
	CancelUnpaidOrder(ctx context.Context) (nAffected int64, err error)
}

type orderRepository struct {
	postgres postgres.PostgresClient
}

func NewOrderRepository(postgres postgres.PostgresClient) OrderRepository {
	return &orderRepository{postgres: postgres}
}

func (repo *orderRepository) Create(ctx context.Context, orders []*model.Order) (baseOrderID int64, OrderID int64, err error) {
	query, args := repo.orderMenusInsertQuery(orders)
	err = repo.postgres.QueryRowContext(ctx, query, args...).Scan(&baseOrderID, &OrderID)
	if err != nil {
		err = fmt.Errorf("repository.ownerRepository.Create: %w", err)
		return 0, 0, err
	}

	return baseOrderID, OrderID, nil
}

func (repo *orderRepository) CancelUnpaidOrder(ctx context.Context) (int64, error) {
	res, err := repo.postgres.ExecContext(ctx, updateOrderStatusToCancelled)
	if err != nil {
		err = fmt.Errorf("repository.ownerRepository.CancelUnpaidOrder: %w", err)
		return 0, err
	}

	nAffected, err := res.RowsAffected()
	if err != nil {
		err = fmt.Errorf("repository.ownerRepository.CancelUnpaidOrder: %w", err)
		return 0, err
	}
	return nAffected, nil
}

func (repo *orderRepository) ConfirmPayment(ctx context.Context, email string) (nAffected int64, errNoRow error, err error) {
	res, err := repo.postgres.ExecContext(ctx, confirmPaymentViaEmail, email)
	if err != nil {
		err = fmt.Errorf("repository.ownerRepository.ConfirmPayment: %w", err)
		return 0, nil, err
	}

	nAffected, err = res.RowsAffected()
	if err == nil && nAffected == 0 {
		err = fmt.Errorf("repository.ownerRepository.ConfirmPayment: %w", sql.ErrNoRows)
		return 0, err, nil
	}

	if err != nil {
		err = fmt.Errorf("repository.ownerRepository.ConfirmPayment: %w", err)
		return 0, nil, err
	}

	return nAffected, nil, nil

}

func (repo *orderRepository) Search(ctx context.Context, order model.OrderQuery) (orders []*model.Order, errNoRow error, err error) {
	query, args := dynamicSearchOrderQuery(&order)
	fmt.Println("query & args: ", query, args)
	rows, err := repo.postgres.QueryContext(ctx, query, args...)
	orders = make([]*model.Order, 0)

	if err != nil {
		err = fmt.Errorf("repository.ownerRepository.Search: %w", err)
		return nil, nil, err
	}

	for rows.Next() {
		row := new(model.Order)
		sliceOfPointerOrderedOrder, err := repo.scanSearchOrderColumn(rows, row) // will return slice of pointer of row field (model.Order)
		if err != nil {
			err = fmt.Errorf("repository.ownerRepository.Search: %w", err)
			return nil, nil, err
		}

		err = rows.Scan(sliceOfPointerOrderedOrder...) // will be mutate row value (side effect)
		if err != nil {
			err = fmt.Errorf("repository.ownerRepository.Search: %w", err)
			return nil, nil, err
		}

		orders = append(orders, row)
	}
	// for decision reason see ./owner.go
	if !rows.Next() && err == nil && len(orders) == 0 {
		err = fmt.Errorf("repository.menuRepository.Search: %w", sql.ErrNoRows)
		return nil, err, nil
	}

	if err != nil {
		err = fmt.Errorf("repository.menuRepository.Search: %w", err)
		return nil, nil, err
	}

	return orders, nil, rows.Close()
}

func (repo *orderRepository) orderMenusInsertQuery(values []*model.Order) (string, []interface{}) {
	if len(values) == 0 {
		return "", []interface{}{}
	}
	stmt := `INSERT INTO "order"(customer_email, menu_id, menu_name, price, qty, status) VALUES %s RETURNING base_order_id, order_id`
	nCols := 6
	valuesStmt := make([]string, 0, len(values))
	args := make([]interface{}, 0, len(values))
	nRowArgs := 0 // start with zero for easier calculation
	for _, val := range values {
		valuesStmt = append(valuesStmt, fmt.Sprintf(
			`($%d, $%d, $%d, $%d, $%d, $%d)`, ((nRowArgs*nCols)+1), ((nRowArgs*nCols)+2),
			((nRowArgs*nCols)+3), ((nRowArgs*nCols)+4), ((nRowArgs*nCols)+5), ((nRowArgs*nCols)+6)))
		nRowArgs += 1

		args = append(args, val.CustomerEmail)
		args = append(args, val.MenuID)
		args = append(args, val.MenuName)
		args = append(args, val.Price)
		args = append(args, val.Qty)
		args = append(args, val.Status)
	}
	stmt = fmt.Sprintf(stmt, strings.Join(valuesStmt, ","))

	return stmt, args
}

func (repo *orderRepository) scanSearchOrderColumn(rows *sql.Rows, order *model.Order) (toScanValue []interface{}, err error) {
	cols, err := rows.Columns()
	if err != nil {
		return []interface{}{}, err
	}
	fmt.Println("cols: ", cols)

	toScanValue = make([]interface{}, 0, len(cols))

	for _, col := range cols {
		switch col {
		case "base_order_id":
			toScanValue = append(toScanValue, &order.BaseOrderID)
		case "customer_email":
			toScanValue = append(toScanValue, &order.CustomerEmail)
		case "menu_name":
			toScanValue = append(toScanValue, &order.MenuName)
		case "order_id":
			toScanValue = append(toScanValue, &order.OrderID)
		case "menu_id":
			toScanValue = append(toScanValue, &order.MenuID)
		case "price":
			toScanValue = append(toScanValue, &order.Price)
		case "qty":
			toScanValue = append(toScanValue, &order.Qty)
		case "status":
			toScanValue = append(toScanValue, &order.Status)
		case "created_at":
			toScanValue = append(toScanValue, &order.CreatedAt)
		case "updated_at":
			toScanValue = append(toScanValue, &order.UpdatedAt)
		}
	}
	fmt.Println("toScanValue: ", toScanValue)
	return toScanValue, nil
}
