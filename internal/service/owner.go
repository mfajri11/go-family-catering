package service

import (
	"context"
	"database/sql"
	"errors"
	"family-catering/internal/model"
	"family-catering/internal/repository"
	"family-catering/pkg/apperrors"
	"family-catering/pkg/consts"
	"fmt"

	utils "family-catering/pkg/utils"
)

type OwnerService interface {
	Create(ctx context.Context, req model.CreateOwnerRequest) (*model.CreateOwnerResponse, error)
	Get(ctx context.Context, id int64) (resp *model.GetOwnerResponse, err error)
	List(ctx context.Context, limit, offset int) (resp []*model.GetOwnerResponse, err error)
	Update(ctx context.Context, id int64, req model.UpdateOwnerRequest) (resp *model.UpdateOwnerResponse, err error)
	Delete(ctx context.Context, id int64) (nAffected int64, err error)
	ResetPasswordByEmail(ctx context.Context, passwordResetID string, req model.ResetPasswordRequest) error
	ResetPasswordByID(ctx context.Context, id int64, req model.ResetPasswordRequest) error
	UpdateEmailByID(ctx context.Context, id int64, req model.UpdateEmailRequest) error
}

type ownerService struct {
	ownerRepo repository.OwnerRepository
}

func NewOwnerService(ownerRepo repository.OwnerRepository) OwnerService {
	return &ownerService{ownerRepo: ownerRepo}
}

func (svc *ownerService) Create(ctx context.Context, req model.CreateOwnerRequest) (*model.CreateOwnerResponse, error) {

	// validating request see internal/model for validation criteria which defined at tag for every request struct
	err := utils.ValidateRequest(&req)
	if err == apperrors.ErrRequiredParam {
		err = fmt.Errorf("service.ownerService.Create: %w", err)
		return nil, apperrors.WrapError(err, apperrors.ErrFieldValidationRequired, "")
	}
	if err != nil {
		err = fmt.Errorf("service.ownerService.Create: %w", err)
		return nil, apperrors.WrapError(err, apperrors.ErrFieldValidation, "")
	}

	// check whether user already registered or not
	ownerFromEmail, errNoRow, err := svc.ownerRepo.GetByEmail(ctx, req.Email)
	// when errNoRow nil means there at least one record/row with that email that already been registered
	if err == nil && errNoRow == nil && ownerFromEmail != nil {
		err = fmt.Errorf("service.ownerService.Create: email already registered")
		return nil, apperrors.WrapError(err, apperrors.ErrEmailRegistered, "")
	}

	if err != nil {
		err = fmt.Errorf("service.ownerService.Create: %w", err)
		return nil, err
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if !errors.Is(err, nil) {
		err = fmt.Errorf("service.ownerService.Create: %w", err)
		return nil, err
	}

	owner := model.Owner{
		Name:        req.Name,
		Email:       req.Email,
		Password:    hashedPassword,
		PhoneNumber: req.PhoneNumber,
	}

	id, err := svc.ownerRepo.Create(ctx, owner)
	if err != nil {
		err = fmt.Errorf("service.ownerService.Create: %w", err)
		return nil, err
	}

	ownerResp := newOwnerCreateResponse(&owner)
	ownerResp.Id = id
	return ownerResp, nil
}

func (svc *ownerService) Get(ctx context.Context, id int64) (*model.GetOwnerResponse, error) {
	owner, errNoRow, err := svc.ownerRepo.Get(ctx, id)
	if errNoRow != nil {
		errNoRow = fmt.Errorf("service.ownerService.Get: %w", errNoRow)
		return nil, apperrors.WrapError(errNoRow, apperrors.ErrNotFound, "")
	}

	if err != nil {
		err = fmt.Errorf("service.ownerService.Get: %w", err)
		return nil, err
	}

	return newOwnerResponse(owner), err
}

func (svc *ownerService) List(ctx context.Context, limit, offset int) ([]*model.GetOwnerResponse, error) {

	owners, errNoRow, err := svc.ownerRepo.List(ctx, limit, offset)
	if errNoRow != nil {
		return []*model.GetOwnerResponse{}, nil
	}
	if err != nil {
		err = fmt.Errorf("service.ownerService.List: %w", err)
		return nil, err
	}

	return newListOwnersResponse(owners), nil
}

func (svc *ownerService) Update(ctx context.Context, id int64, req model.UpdateOwnerRequest) (*model.UpdateOwnerResponse, error) {

	// validate authorization
	token, ok := utils.ValueContext(ctx, consts.CtxKeyAuthorization).(string) // ? does use generic could be reduce this repetition?
	if !ok {
		err := fmt.Errorf("service.ownerService.Update: invalid auth token type want string got %T", token)
		return nil, apperrors.WrapError(err, apperrors.ErrAuth, "invalid auth token type")
	}
	_, err := utils.ValidateToken(token)
	if !errors.Is(err, nil) {
		err = fmt.Errorf("service.ownerService.Update: %w", err)
		return nil, apperrors.WrapError(err, apperrors.ErrAuth, "")
	}

	// validating request
	err = utils.ValidateRequest(&req)
	if errors.Is(err, apperrors.ErrRequiredParam) {
		err = fmt.Errorf("service.ownerService.Update: %w", err)
		return nil, apperrors.WrapError(err, apperrors.ErrFieldValidationRequired, "")
	}
	if !errors.Is(err, nil) {
		err = fmt.Errorf("service.ownerService.Update: %w", err)
		return nil, apperrors.WrapError(err, apperrors.ErrFieldValidation, "")
	}

	owner := model.Owner{
		Name:        req.Name,
		PhoneNumber: req.PhoneNumber,
	}

	if req.DateOfBirth != "" {
		owner.DateOfBirth = sql.NullString{String: req.DateOfBirth, Valid: true}
	}

	owner.Id = id
	_, errNoRow, err := svc.ownerRepo.Update(ctx, owner)
	if errNoRow != nil {
		errNoRow = fmt.Errorf("service.ownerService.Update: %w", errNoRow)
		return nil, apperrors.WrapError(errNoRow, apperrors.ErrNotFound, "")
	}

	if err != nil {
		err = fmt.Errorf("service.ownerService.Update: %w", err)
		return nil, err
	}

	return newOwnerResponse(&owner), nil
}

func (svc *ownerService) Delete(ctx context.Context, id int64) (int64, error) {
	// validate authorization (owner can only delete him/her-self)
	token, ok := utils.ValueContext(ctx, consts.CtxKeyAuthorization).(string)
	if !ok {
		err := fmt.Errorf("service.ownerService.Delete: invalid auth token type want string got %T", token)
		return 0, apperrors.WrapError(err, apperrors.ErrAuth, "invalid auth token type")
	}
	session, ok := utils.ValueContext(ctx, consts.CtxKeySession).(*model.AuthSessionResponse)
	if !ok {
		err := fmt.Errorf("service.ownerService.Delete: invalid session")
		return 0, apperrors.WrapError(err, apperrors.ErrAuth, "")
	}
	_, err := utils.ValidateToken(token)

	if !errors.Is(err, nil) {
		err = fmt.Errorf("service.ownerService.Delete: %w", err)
		return 0, apperrors.WrapError(err, apperrors.ErrAuth, "")
	}
	if session.OwnerID != id || !session.Valid {
		err = fmt.Errorf("service.ownerService.Delete: invalid session or mismatch id")
		return 0, apperrors.WrapError(err, apperrors.ErrAuth, "")
	}

	nAffected, errNoRow, err := svc.ownerRepo.Delete(ctx, id)
	if errNoRow != nil {
		errNoRow = fmt.Errorf("service.ownerService.Delete: %w", errNoRow)
		return 0, apperrors.WrapError(errNoRow, apperrors.ErrNotFound, "")
	}
	if err != nil {
		err = fmt.Errorf("service.ownerService.Delete: %w", err)
		return 0, err
	}
	return nAffected, nil
}

func (svc *ownerService) ResetPasswordByEmail(ctx context.Context, passwordResetID string, req model.ResetPasswordRequest) error {
	// validate authorization
	token, ok := utils.ValueContext(ctx, consts.CtxKeyAuthorization).(string)
	if !ok {
		err := fmt.Errorf("service.ownerService.ResetPasswordByEmail: invalid auth token type want string got %T", token)
		return apperrors.WrapError(err, apperrors.ErrAuth, "invalid auth token type")
	}

	payload, err := utils.ValidateToken(token)
	if !errors.Is(err, nil) {
		err = fmt.Errorf("service.ownerService.ResetPasswordByEmail: %w", err)
		return apperrors.WrapError(err, apperrors.ErrAuth, "")
	}

	if !payload.IsForResetPassword() || (payload.Id != passwordResetID) || (req.Password != req.PasswordConfirm) {
		err := fmt.Errorf("service.ownerService.ResetPasswordByEmail: invalid jwt claim or request value")
		return apperrors.WrapError(err, apperrors.ErrAuth, "")
	}
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		err = fmt.Errorf("service.ownerService.ResetPasswordByEmail: %w", err)
		return err
	}
	_, errNoRow, err := svc.ownerRepo.UpdatePasswordByEmail(context.Background(), payload.Email, hashedPassword)
	if errNoRow != nil {
		errNoRow = fmt.Errorf("service.ownerService.ResetPasswordByEmail: %w", errNoRow)
		return apperrors.WrapError(errNoRow, apperrors.ErrNotFound, "")
	}
	if err != nil {
		err = fmt.Errorf("service.ownerService.ResetPasswordByEmail: %w", err)
		return err
	}

	return nil

}

func (svc *ownerService) ResetPasswordByID(ctx context.Context, id int64, req model.ResetPasswordRequest) error {
	// must in logged state
	session, ok := utils.ValueContext(ctx, consts.CtxKeySession).(*model.AuthSessionResponse)
	if !ok {
		err := fmt.Errorf("service.ownerService.ResetPasswordByID: invalid session")
		return apperrors.WrapError(err, apperrors.ErrAuth, "")
	}
	// validate authorization
	token := utils.ValueContext(ctx, consts.CtxKeyAuthorization).(string)
	if !ok {
		err := fmt.Errorf("service.ownerService.ResetPasswordById: invalid auth token type want string got %T", token)
		return apperrors.WrapError(err, apperrors.ErrAuth, "invalid auth token type")
	}
	_, err := utils.ValidateToken(token)
	if !errors.Is(err, nil) {
		err = fmt.Errorf("service.ownerService.ResetPasswordByID: %w", err)
		return apperrors.WrapError(err, apperrors.ErrAuth, "")
	}
	if !session.Valid || id != session.OwnerID || req.Password != req.PasswordConfirm {
		err = fmt.Errorf("service.ownerService.ResetPasswordByID: invalid session or request")
		return apperrors.WrapError(err, apperrors.ErrAuth, "")
	}
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		err = fmt.Errorf("service.ownerService.ResetPasswordByID: %w", err)
		return err
	}
	_, errNoRow, err := svc.ownerRepo.UpdatePasswordByID(ctx, id, hashedPassword)
	if errNoRow != nil {
		err = fmt.Errorf("service.ownerService.ResetPasswordByID: %w", err)
		return apperrors.WrapError(err, apperrors.ErrNotFound, "")
	}
	if err != nil {
		err = fmt.Errorf("service.ownerService.ResetPasswordByID: %w", err)
		return err
	}
	return nil
}

func (svc *ownerService) UpdateEmailByID(ctx context.Context, id int64, req model.UpdateEmailRequest) error {
	// validate session (must in logged state)
	session, ok := utils.ValueContext(ctx, consts.CtxKeySession).(*model.AuthSessionResponse)
	if !ok {
		err := fmt.Errorf("service.ownerService.UpdateEmailByID: invalid session")
		return apperrors.WrapError(err, apperrors.ErrAuth, "")
	}
	// validate authorization
	token := utils.ValueContext(ctx, consts.CtxKeyAuthorization).(string)
	if !ok {
		err := fmt.Errorf("service.ownerService.UpdateEmailByID: invalid auth token type want string got %T", token)
		return apperrors.WrapError(err, apperrors.ErrAuth, "invalid auth token type")
	}
	_, err := utils.ValidateToken(token)
	if !errors.Is(err, nil) {
		err = fmt.Errorf("service.ownerService.ResetPasswordByEmail: %w", err)
		return apperrors.WrapError(err, apperrors.ErrAuth, "")
	}
	if !session.Valid || (id != session.OwnerID) {
		err = fmt.Errorf("service.ownerService.ResetPasswordByEmail: invalid session or claims")
		return apperrors.WrapError(err, apperrors.ErrAuth, "")
	}

	_, errNoRow, err := svc.ownerRepo.UpdateEmailByID(ctx, id, req.Email)
	if errNoRow != nil {
		errNoRow = fmt.Errorf("service.ownerService.ResetPasswordByID: %w", errNoRow)
		return apperrors.WrapError(errNoRow, apperrors.ErrNotFound, "")
	}
	if err != nil {
		err = fmt.Errorf("service.ownerService.ResetPasswordByID: %w", err)
		return err
	}

	return nil
}
