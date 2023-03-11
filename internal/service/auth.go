package service

import (
	"errors"
	"family-catering/config"
	"family-catering/internal/model"
	"family-catering/internal/repository"
	"family-catering/pkg/apperrors"
	"family-catering/pkg/consts"
	"family-catering/pkg/utils"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/context"
)

type AuthService interface {
	Login(ctx context.Context, req model.AuthLoginRequest) (resp *model.AuthLoginResponse, err error)
	Logout(ctx context.Context, req model.AuthLogoutRequest) error
	ForgotPassword(ctx context.Context, req model.AuthForgotPasswordRequest) (token string, err error)
	Session(ctx context.Context, sid string) (resp *model.AuthSessionResponse, err error)
	RenewAccessToken(ctx context.Context) (resp *model.AuthRenewAccessTokenResponse, err error)
}

type authService struct {
	ownerRepo repository.OwnerRepository
	authRepo  repository.AuthRepository
	mailer    Mailer
}

func NewAuthService(ownerRepo repository.OwnerRepository, authRepo repository.AuthRepository, mailer Mailer) AuthService {
	return &authService{ownerRepo: ownerRepo, authRepo: authRepo, mailer: mailer}
}

func (svc *authService) Login(ctx context.Context, req model.AuthLoginRequest) (resp *model.AuthLoginResponse, err error) {
	// validating request
	err = utils.ValidateRequest(&req)
	if err == apperrors.ErrRequiredParam {
		return nil, apperrors.WrapError(err, apperrors.ErrFieldValidationRequired, "")
	}
	if err != nil {
		return nil, apperrors.WrapError(err, apperrors.ErrFieldValidation, "")
	}

	// get owner to compare given password with stored password
	owner, errNoRow, err := svc.ownerRepo.GetByEmail(ctx, req.Email)
	if errNoRow != nil && err == nil && owner != nil {
		return nil, apperrors.WrapError(errNoRow, apperrors.ErrNotFound, "")
	}
	if err != nil {
		return nil, fmt.Errorf("service.authRepository.Login: %w", err)
	}

	// validate password
	err = utils.ValidatePassword(req.Password, owner.Password)
	if err != nil {
		return nil, apperrors.WrapError(err, apperrors.ErrAuth, "error wrong password")
	}

	// check if owner has sid associate with its email (means already logged in in another device)
	sid, err := svc.authRepo.GetSessionIDByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("service.authRepository.Login: %w", err)
	}

	// owner has active session associated with its email
	if err == nil && sid != "" {
		autLoginResponse := model.AuthLoginResponse{
			AccessToken:  "",
			RefreshToken: "",
			SID:          sid,
		}
		return &autLoginResponse, nil
	}

	autLogin := model.Auth{}
	autLogin.Email = req.Email
	autLogin.OwnerID = owner.Id
	// generate session id used to indicate wether the owner authenticated or not
	sid = uuid.New().String()
	autLogin.SID = sid

	// generate access token (JWT)
	// owner which has access token assumed is logged (authenticated)
	accessToken, err := utils.GenerateToken(svc.authRepo.AccessTokenTTL(), "", "") // id only use in forgot password
	if err != nil {
		return nil, fmt.Errorf("service.authRepository.Login: %w", err)
	}

	// generate refresh token (JWT)
	autLogin.Jti = uuid.New().String()
	refreshToken, err := utils.GenerateToken(svc.authRepo.RefreshTokenTTL(), autLogin.Jti, "") // 60 days
	if err != nil {
		return nil, fmt.Errorf("service.authRepository.Login: %w", err)
	}

	autLogin.RefreshToken = refreshToken

	// store userid, sid & access token to redis
	// store userid, sid, refresh token to postgres
	err = svc.authRepo.Login(ctx, autLogin)
	if err != nil {
		return nil, fmt.Errorf("service.authRepository.Login: %w", err)
	}

	resp = &model.AuthLoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SID:          sid,
	}

	// svc.mailer.SendEmailNotifyLogin([]string{owner.Email}, "", "Login Family-Catering", owner.Name)

	return resp, nil
}

func (svc *authService) Logout(ctx context.Context, req model.AuthLogoutRequest) error {
	// check sid
	sid, ok := utils.ValueContext(ctx, consts.CtxKeySID).(string)
	if !ok || sid == "" {
		err := errors.New("service.authService.Logout: missing session id")
		return apperrors.WrapError(err, apperrors.ErrAuth, "missing session id")
	}
	req.SID = sid
	// validate authorization
	token, ok := utils.ValueContext(ctx, consts.CtxKeyAuthorization).(string)
	if !ok {
		err := fmt.Errorf("service.authService.Logout: invalid auth token type want string got %T", token)
		return apperrors.WrapError(err, apperrors.ErrAuth, "invalid auth token type")
	}
	_, err := utils.ValidateToken(token)
	if !errors.Is(err, nil) {
		err = fmt.Errorf("service.authService.Logout: %w", err)
		return apperrors.WrapError(err, apperrors.ErrAuth, "invalid Authorization value")
	}

	// validate request
	err = utils.ValidateRequest(&req)
	if err == apperrors.ErrRequiredParam {
		err = fmt.Errorf("service.authService.Logout: %w", err)
		return apperrors.WrapError(err, apperrors.ErrFieldValidationRequired, "")
	}
	if err != nil {
		err = fmt.Errorf("service.authService.Logout: %w", err)
		return apperrors.WrapError(err, apperrors.ErrFieldValidation, "")
	}

	// get session object from given sid
	ownerSession, errNoRow, err := svc.authRepo.Session(ctx, req.SID)
	if errNoRow != nil {
		return fmt.Errorf("service.authService.Logout: %w", errNoRow)
	}
	if err != nil { // \!ok mean no entry associated with given sid
		return fmt.Errorf("service.authService.Logout: %w", err)
	}

	// check session valid or not
	if !ownerSession.Valid {
		return fmt.Errorf("service.authService.Logout: err session is not valid")
	}
	// get owner associated with given sid to compare given password (from request) with stored password from owner object stored
	owner, errNoRow, err := svc.ownerRepo.Get(ctx, ownerSession.OwnerID)
	if err != nil {
		err = fmt.Errorf("service.authService.Logout: %w", err)
		return err
	}
	if errNoRow != nil {
		errNoRow = fmt.Errorf("service.authService.Logout: %w", errNoRow)
		return errNoRow
	}

	// compare password from request & database
	err = utils.ValidatePassword(req.Password, owner.Password)
	if err != nil {
		err = fmt.Errorf("service.authService.Logout: %w", err)
		return err
	}

	// if valid password delete sid (both from redis & postgres)
	err = svc.authRepo.DeleteSession(ctx, req.SID)
	if err != nil {
		err = fmt.Errorf("service.authService.Logout: %w", err)
		return err
	}

	return nil
}

func (svc *authService) ForgotPassword(ctx context.Context, req model.AuthForgotPasswordRequest) (string, error) {
	// validate request
	err := utils.ValidateRequest(&req)
	if err == apperrors.ErrRequiredParam {
		err = fmt.Errorf("service.authRepository.ForgotPassword: %w", err)
		return "", apperrors.WrapError(err, apperrors.ErrFieldValidationRequired, "")
	}
	if err != nil {
		err = fmt.Errorf("service.authRepository.ForgotPassword: %w", err)
		return "", apperrors.WrapError(err, apperrors.ErrFieldValidation, "")
	}

	owner, errNoRow, err := svc.ownerRepo.GetByEmail(ctx, req.Email)
	if errNoRow != nil {
		err = fmt.Errorf("service.authRepository.ForgotPassword: %w", err)
		return "", apperrors.WrapError(err, apperrors.ErrNotFound, "email not found")
	}
	if err != nil {
		err = fmt.Errorf("service.authRepository.ForgotPassword: %w", err)
		return "", err
	}
	uid := uuid.New()
	id := uid.String()
	token, err := utils.GenerateToken(svc.authRepo.AccessTokenTTL(), id, owner.Email)
	if err != nil {
		err = fmt.Errorf("service.authRepository.ForgotPassword: %w", err)
		return "", err
	}

	// email must be sent to the request's email
	requestLink := fmt.Sprintf("https://%s/api/v1/owner/reset-password/%s", config.Cfg().Server.Addr(), id)
	err = svc.mailer.SendEMailForgotPassword([]string{owner.Email}, "", "Reset Password", owner.Name, requestLink)
	if err != nil {
		err = fmt.Errorf("service.authRepository.ForgotPassword: %w", err)
		return "", err
	}

	return token, nil
}

func (svc *authService) RenewAccessToken(ctx context.Context) (*model.AuthRenewAccessTokenResponse, error) {
	Session, okSession := utils.ValueContext(ctx, consts.CtxKeySession).(*model.AuthSessionResponse)
	token, ok := utils.ValueContext(ctx, consts.CtxKeyAuthorization).(string)
	if !ok || !okSession {
		err := fmt.Errorf("service.authService.RenewAccessToken: invalid auth token type want string got %T", token)
		return nil, apperrors.WrapError(err, apperrors.ErrAuth, "invalid auth token type")
	}
	payload, err := utils.ValidateToken(token)
	if !errors.Is(err, nil) {
		err = fmt.Errorf("service.authRepository.RenewAccessToken: %w", err)
		return nil, err
	}

	if !ok || !Session.Valid || (Session.Jti != payload.Id) || !payload.IsForRefreshToken() {
		err = fmt.Errorf("service.authRepository.RenewAccessToken: invalid session or claims")
		return nil, apperrors.WrapError(err, apperrors.ErrAuth, "")
	}

	accessToken, err := utils.GenerateToken(svc.authRepo.AccessTokenTTL(), "", "")
	if err != nil {
		err = fmt.Errorf("service.authRepository.RenewAccessToken: %w", err)
		return nil, err
	}

	expiredAt := time.Now().Add(svc.authRepo.AccessTokenTTL()).Format(time.RFC3339)

	return &model.AuthRenewAccessTokenResponse{AccessToken: accessToken, ExpiredAt: expiredAt}, nil

}

func (svc *authService) Session(ctx context.Context, sid string) (*model.AuthSessionResponse, error) {
	ownerSession, errNoRow, err := svc.authRepo.Session(ctx, sid)
	if err != nil {
		return nil, fmt.Errorf("service.authService.Session: %w", err)
	}
	if errNoRow != nil {
		return nil, nil
	}

	session := model.AuthSessionResponse{
		SID:     ownerSession.SID,
		OwnerID: ownerSession.OwnerID,
		Jti:     ownerSession.Jti,
		Valid:   ownerSession.Valid,
	}

	return &session, nil
}
