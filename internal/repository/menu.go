package repository

import (
	"context"
	"database/sql"
	"family-catering/internal/model"
	"family-catering/pkg/db/postgres"
	"fmt"
)

type MenuRepository interface {
	GetByID(ctx context.Context, id int64) (menu *model.Menu, errNoRow error, err error)
	GetByName(ctx context.Context, name string) (menu *model.Menu, errNoRow error, err error)
	List(ctx context.Context, limit, offset int) (menus []*model.Menu, errNoRow error, err error)
	Update(ctx context.Context, menu model.Menu) (nAffected int64, errNoRow error, err error)
	Create(ctx context.Context, menu model.Menu) (id int64, err error)
	Delete(ctx context.Context, id int64) (nAffected int64, errNoRow error, err error)
	Search(ctx context.Context, menu model.MenuQuery) (menus []*model.Menu, errNoRow error, err error)
}

type menuRepository struct {
	postgres postgres.PostgresClient
}

func NewMenuRepository(postgres postgres.PostgresClient) MenuRepository {
	return &menuRepository{postgres: postgres}
}

func (repo *menuRepository) GetByID(ctx context.Context, id int64) (menu *model.Menu, errNoRow error, err error) {

	menu = &model.Menu{}
	err = repo.postgres.
		QueryRowContext(ctx, getMenuByID, id).
		Scan(
			&menu.ID,
			&menu.Name,
			&menu.Price,
			&menu.Categories,
		)

	if err == sql.ErrNoRows {
		err = fmt.Errorf("repository.menuRepository.GetByID: %w", err)
		return nil, err, nil
	}

	if err != nil {
		err = fmt.Errorf("repository.menuRepository.GetByID: %w", err)
		return nil, nil, err
	}

	return menu, nil, nil
}

func (repo *menuRepository) GetByName(ctx context.Context, name string) (menu *model.Menu, errNoRow error, err error) {

	menu = &model.Menu{}
	err = repo.postgres.
		QueryRowContext(ctx, getMenuByName, name).
		Scan(
			&menu.ID,
			&menu.Name,
			&menu.Price,
			&menu.Categories,
		)

	if err == sql.ErrNoRows {
		err = fmt.Errorf("repository.menuRepository.GetByName: %w", err)
		return nil, err, nil
	}

	if err != nil {
		err = fmt.Errorf("repository.menuRepository.GetByName: %w", err)
		return nil, nil, err
	}

	return menu, nil, nil
}

func (repo *menuRepository) List(ctx context.Context, limit, offset int) (menus []*model.Menu, errNoRow error, err error) {
	rows, err := repo.postgres.QueryContext(ctx, listMenu, limit, offset)
	if err != nil {
		err = fmt.Errorf("repository.menuRepository.List: %w", err)
		return nil, nil, err
	}

	defer rows.Close()

	for rows.Next() {
		menu := new(model.Menu)
		err = rows.Scan(
			&menu.ID,
			&menu.Name,
			&menu.Price,
			&menu.Categories,
		)

		if err != nil {
			err = fmt.Errorf("repository.menuRepository.List: %w", err)
			return nil, nil, err
		}

		menus = append(menus, menu)
	}

	err = rows.Err()

	// for decision reason see ./owner.go
	if !rows.Next() && err == nil && len(menus) == 0 {
		err = fmt.Errorf("repository.menuRepository.List: %w", sql.ErrNoRows)
		return nil, err, nil
	}

	if err != nil {
		err = fmt.Errorf("repository.menuRepository.List: %w", err)
		return nil, nil, err
	}

	return menus, nil, rows.Close()
}

func (repo *menuRepository) Create(ctx context.Context, menu model.Menu) (id int64, err error) {
	err = repo.postgres.QueryRowContext(ctx, createMenu, menu.Name, menu.Price, menu.Categories).Scan(&id)
	if err != nil {
		err = fmt.Errorf("repository.menuRepository.Create: %w", err)
		return 0, err
	}
	return id, nil
}

func (repo *menuRepository) Update(ctx context.Context, menu model.Menu) (nAffected int64, errNoRow error, err error) {

	res, err := repo.postgres.ExecContext(ctx, updateMenuByID, menu.ID, menu.Name, menu.Price, menu.Categories)

	if err != nil {
		err = fmt.Errorf("repository.menuRepository.Update: %w", err)
		return 0, nil, err
	}

	nAffected, err = res.RowsAffected()
	if err != nil {
		err = fmt.Errorf("repository.menuRepository.Update: %w", err)
		return 0, nil, err
	}

	if err == nil && nAffected == 0 {
		return 0, fmt.Errorf("repository.menuRepository.Update: %w", sql.ErrNoRows), nil
	}

	return nAffected, nil, nil
}

func (repo *menuRepository) Delete(ctx context.Context, id int64) (nAffected int64, errNoRow error, err error) {
	res, err := repo.postgres.ExecContext(ctx, deleteMenuByID, id)

	if err != nil {
		err = fmt.Errorf("repository.menuRepository.Delete: %w", err)
		return 0, nil, err
	}

	nAffected, err = res.RowsAffected()
	if err != nil {
		err = fmt.Errorf("repository.menuRepository.Delete: %w", err)
		return 0, nil, err
	}

	if err == nil && nAffected == 0 {
		return 0, fmt.Errorf("repository.menuRepository.Delete: %w", sql.ErrNoRows), nil
	}

	return nAffected, nil, nil
}

func (repo *menuRepository) Search(ctx context.Context, menu model.MenuQuery) (menus []*model.Menu, errNoRow error, err error) {
	menus = make([]*model.Menu, 0)
	query, args := menuDynamicSearchQuery(menu)
	rows, err := repo.postgres.QueryContext(ctx, query, args...)
	if err != nil {
		err = fmt.Errorf("repository.menuRepository.Search: %w", err)
		return nil, nil, err
	}
	for rows.Next() {
		row := new(model.Menu)
		sliceOfPointerOrderedRow, err := repo.scanSearchMenuColumnOrder(rows, row) // will return slice of pointer of row field (model.Menu)
		if err != nil {
			err = fmt.Errorf("repository.menuRepository.Search: %w", err)
			return nil, nil, err
		}

		err = rows.Scan(sliceOfPointerOrderedRow...) // will be mutate row value (side effect)
		if err != nil {
			err = fmt.Errorf("repository.menuRepository.Search: %w", err)
			return nil, nil, err
		}

		menus = append(menus, row)
	}
	err = rows.Err()
	// for decision reason see ./owner.go
	if !rows.Next() && err == nil && len(menus) == 0 {
		err = fmt.Errorf("repository.menuRepository.Search: %w", sql.ErrNoRows)
		return nil, err, nil
	}

	if err != nil {
		err = fmt.Errorf("repository.menuRepository.Search: %w", err)
		return nil, nil, err
	}

	return menus, nil, rows.Close()
}

func (repo *menuRepository) scanSearchMenuColumnOrder(rows *sql.Rows, menu *model.Menu) (toScanValue []interface{}, err error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	res := []interface{}{}
	for _, col := range cols {
		switch col {
		case "name":
			res = append(res, &menu.Name)
		case "id":
			res = append(res, &menu.ID)
		case "price":
			res = append(res, &menu.Price)
		case "categories":
			res = append(res, &menu.Categories)
		}
	}
	return res, nil
}
