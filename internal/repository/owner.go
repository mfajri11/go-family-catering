package repository

import (
	"context"
	"database/sql"
	"family-catering/internal/model"
	"family-catering/pkg/db/postgres"
	"fmt"
)

type OwnerRepository interface {
	Create(ctx context.Context, owner model.Owner) (id int64, err error)
	Get(ctx context.Context, id int64) (owner *model.Owner, errNoRow error, err error)
	GetByEmail(ctx context.Context, email string) (owner *model.Owner, errNoRow error, err error)
	List(ctx context.Context, limit, offset int) (owners []*model.Owner, errNoRow error, err error)
	Update(ctx context.Context, owner model.Owner) (nAffected int64, errNoRow error, err error)
	Delete(ctx context.Context, id int64) (nAffected int64, errNoRow error, err error)
	UpdatePasswordByEmail(ctx context.Context, email, password string) (nAffected int64, errNoRow error, err error)
	UpdatePasswordByID(ctx context.Context, id int64, password string) (nAffected int64, errNoRow error, err error)
	UpdateEmailByID(ctx context.Context, id int64, email string) (nAffected int64, errNoRow error, err error)
}

type ownerRepository struct {
	postgres postgres.PostgresClient
}

func NewOwnerRepository(postgreClient postgres.PostgresClient) OwnerRepository {
	return &ownerRepository{postgres: postgreClient}
}

func (repo *ownerRepository) Create(ctx context.Context, owner model.Owner) (int64, error) {
	var id int64
	err := repo.postgres.QueryRowContext(ctx, createOwner, owner.Name, owner.Email, owner.Password, owner.PhoneNumber).Scan(&id)
	if err != nil {
		err := fmt.Errorf("repository.ownerRepository.Create: %w", err)
		return 0, err
	}

	return id, nil
}

func (repo *ownerRepository) Update(ctx context.Context, owner model.Owner) (int64, error, error) {
	var (
		err       error
		nAffected int64
	)
	res, err := repo.postgres.ExecContext(ctx, updateOwner, owner.Id, owner.Name, owner.PhoneNumber, owner.DateOfBirth.String)
	if err != nil {
		err = fmt.Errorf("repository.ownerRepository.Update: %w", err)
		return 0, nil, err

	}

	nAffected, err = res.RowsAffected()
	if err != nil {
		err = fmt.Errorf("repository.ownerRepository.Update: %w", err)
		return 0, nil, err

	}
	if err == nil && nAffected == 0 {
		err = fmt.Errorf("repository.ownerRepository.Update: %w", sql.ErrNoRows)
		return 0, err, nil
	}

	return nAffected, nil, nil
}

func (repo *ownerRepository) GetByEmail(ctx context.Context, email string) (owner *model.Owner, errNoRow error, err error) {
	row := repo.postgres.QueryRowContext(ctx, getOwnerByEmail, email)
	owner = &model.Owner{}
	err = row.Scan(
		&owner.Id,
		&owner.Name,
		&owner.Email,
		&owner.PhoneNumber,
		&owner.DateOfBirth,
		&owner.Password,
	)
	if err == sql.ErrNoRows {
		err = fmt.Errorf("repository.ownerRepository.GetByEmail: %w", err)
		return nil, err, nil
	}

	if err != nil {
		err = fmt.Errorf("repository.ownerRepository.GetByEmail: %w", err)
		return nil, nil, err
	}

	return owner, nil, nil
}

func (repo *ownerRepository) Get(ctx context.Context, id int64) (owner *model.Owner, errNoRow error, err error) {
	row := repo.postgres.QueryRowContext(ctx, getOwner, id)
	owner = &model.Owner{}
	err = row.Scan(
		&owner.Id,
		&owner.Name,
		&owner.Email,
		&owner.PhoneNumber,
		&owner.DateOfBirth,
		&owner.Password,
	)

	if err == sql.ErrNoRows {
		err = fmt.Errorf("repository.ownerRepository.Get: %w", err)
		return nil, err, nil
	}

	if err != nil {
		err = fmt.Errorf("repository.ownerRepository.GetByEmail: %w", err)
		return nil, nil, err
	}

	return owner, nil, nil
}

func (repo *ownerRepository) List(ctx context.Context, limit, offset int) (owners []*model.Owner, errNoRow error, err error) {
	rows, err := repo.postgres.QueryContext(ctx, listOwners, limit, offset)
	if err != nil {
		err = fmt.Errorf("repository.ownerRepository.List: %w", err)
		return nil, nil, err
	}

	defer rows.Close()

	for rows.Next() {
		owner := new(model.Owner)
		err = rows.Scan(
			&owner.Id,
			&owner.Name,
			&owner.Email,
			&owner.PhoneNumber,
			&owner.DateOfBirth,
		)

		if err != nil {
			err = fmt.Errorf("repository.ownerRepository.List: %w", err)
			return nil, nil, err
		}

		owners = append(owners, owner)
	}

	err = rows.Err()

	// based on row.Scan() https://cs.opensource.google/go/go/+/refs/tags/go1.19.5:src/database/sql/sql.go;l=3380
	// which return ErrNoRow when rows.Next() is false and err is nil
	// so the same reason is used + check whether the len of owners is empty
	// if all requirements are satisfied then ErrNoRow is returned from this function
	// this assume no iteration is processed and rows has no entry indicate by false on first rows.Next() called (because of len(owners) is zero)
	if !rows.Next() && err == nil && len(owners) == 0 {
		err = fmt.Errorf("repository.ownerRepository.List: %w", sql.ErrNoRows)
		return nil, err, nil
	}

	if err != nil {
		err = fmt.Errorf("repository.ownerRepository.List: %w", err)
		return nil, nil, err
	}

	return owners, nil, rows.Close()
}

func (repo *ownerRepository) Delete(ctx context.Context, id int64) (int64, error, error) {
	var (
		err       error
		nAffected int64
	)
	res, err := repo.postgres.ExecContext(ctx, deleteOwner, id)
	if err != nil {
		err = fmt.Errorf("repository.ownerRepository.Delete: %w", err)
		return 0, nil, err
	}

	nAffected, err = res.RowsAffected()
	if err == nil && nAffected == 0 {
		err = fmt.Errorf("repository.ownerRepository.Delete: %w", sql.ErrNoRows)
		return 0, err, nil
	}

	if err != nil {
		err = fmt.Errorf("repository.ownerRepository.Delete: %w", err)
		return 0, nil, err
	}

	return nAffected, nil, nil
}

func (repo *ownerRepository) UpdatePasswordByEmail(ctx context.Context, email, password string) (nAffected int64, errNoRow error, err error) {
	res, err := repo.postgres.ExecContext(ctx, updateOwnerPasswordByEmail, email, password)
	if err != nil {
		err = fmt.Errorf("repository.ownerRepository.UpdatePasswordByEmail: %w", err)
		return 0, nil, err
	}

	nAffected, err = res.RowsAffected()
	if err == nil && nAffected == 0 {
		err = fmt.Errorf("repository.ownerRepository.UpdatePasswordByEmail: %w", sql.ErrNoRows)
		return 0, err, nil
	}
	if err != nil {
		err = fmt.Errorf("repository.ownerRepository.UpdatePasswordByEmail: %w", err)
		return 0, nil, err
	}

	return nAffected, nil, nil
}

func (repo *ownerRepository) UpdatePasswordByID(ctx context.Context, id int64, password string) (nAffected int64, errNoRow error, err error) {
	res, err := repo.postgres.ExecContext(ctx, updateOwnerPasswordByID, id, password)
	if err != nil {
		err = fmt.Errorf("repository.ownerRepository.UpdatePasswordByID: %w", err)
		return 0, nil, err
	}

	nAffected, err = res.RowsAffected()
	if err == nil && nAffected == 0 {
		err = fmt.Errorf("repository.ownerRepository.UpdatePasswordByID: %w", sql.ErrNoRows)
		return 0, err, nil
	}
	if err != nil {
		err = fmt.Errorf("repository.ownerRepository.UpdatePasswordByID: %w", err)
		return 0, nil, err
	}

	return nAffected, nil, nil
}

func (repo *ownerRepository) UpdateEmailByID(ctx context.Context, id int64, email string) (nAffected int64, errNoRow error, err error) {
	res, err := repo.postgres.ExecContext(ctx, updateEmailByID, id, email)
	if err != nil {
		err = fmt.Errorf("repository.ownerRepository.UpdatePasswordByID: %w", err)
		return 0, nil, err
	}

	nAffected, err = res.RowsAffected()
	if err == nil && nAffected == 0 {
		err = fmt.Errorf("repository.ownerRepository.UpdatePasswordByID: %w", sql.ErrNoRows)
		return 0, err, nil
	}
	if err != nil {
		err = fmt.Errorf("repository.ownerRepository.UpdatePasswordByID: %w", err)
		return 0, nil, err
	}

	return nAffected, nil, nil
}
