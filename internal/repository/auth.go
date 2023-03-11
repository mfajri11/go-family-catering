package repository

import (
	"context"
	"database/sql"
	"errors"
	"family-catering/config"
	"family-catering/internal/model"
	"family-catering/pkg/db/postgres"
	"family-catering/pkg/db/redis"
	"fmt"
	"time"

	redisV8 "github.com/go-redis/redis/v8"
)

const (
	sessionKeyFormat     string = "sid:%s"
	sessionByEmailFormat string = "email:sid:%s"
)

var (
	accessTokenTTL  = config.Cfg().Web.AccessTokenTTL
	refreshTokenTTL = config.Cfg().Web.RefreshTokenTTL
)

type AuthRepository interface {
	Login(ctx context.Context, authLogin model.Auth) error
	Session(ctx context.Context, sid string) (authLogoutResponse *model.Auth, errNoRow error, err error)
	AccessTokenTTL() time.Duration
	RefreshTokenTTL() time.Duration
	DeleteSession(ctx context.Context, sid string) error
	GetSessionIDByEmail(ctx context.Context, email string) (sessionID string, err error)
}

type authRepository struct {
	postgres        postgres.PostgresClient
	redis           redis.RedisClient
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewAuthRepository(postgres postgres.PostgresClient, redis redis.RedisClient) AuthRepository {
	repo := &authRepository{postgres: postgres, redis: redis}
	repo.accessTokenTTL = accessTokenTTL
	repo.refreshTokenTTL = refreshTokenTTL
	return repo

}

func (repo *authRepository) setRedisSession(ctx context.Context, key string, auth model.Auth) error {
	_, err := repo.redis.Pipelined(ctx, func(p redisV8.Pipeliner) error {
		p.HSet(ctx, key, "owner_id", auth.OwnerID)
		p.HSet(ctx, key, "valid", true)
		p.HSet(ctx, key, "jti", auth.Jti)
		p.HSet(ctx, key, "email", auth.Email)
		p.Expire(ctx, key, repo.refreshTokenTTL)

		p.SetNX(ctx, fmt.Sprintf(sessionByEmailFormat, auth.Email), auth.SID, repo.refreshTokenTTL)
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (repo *authRepository) deleteRedisSession(ctx context.Context, key string) error {

	email := repo.redis.HGet(ctx, key, "email").Val()
	fmt.Println("email: ", email)
	_, err := repo.redis.Pipelined(ctx, func(p redisV8.Pipeliner) error {
		p.HDel(ctx, key, "owner_id")
		p.HDel(ctx, key, "valid")
		p.HDel(ctx, key, "jti")
		p.HDel(ctx, key, "email")
		p.Del(ctx, key)
		p.Del(ctx, fmt.Sprintf(sessionByEmailFormat, email))
		return nil
	})

	fmt.Println("error: ", err)
	if err != nil {
		return err
	}
	return nil
}

func (repo *authRepository) redisSession(ctx context.Context, key string) (*model.Auth, error) {
	session := model.Auth{}
	err := repo.redis.HGetAll(ctx, key).Scan(&session)
	if err != nil {
		return nil, err
	}

	if session == (model.Auth{}) {
		return nil, nil
	}
	return &session, nil
}

func (repo *authRepository) AccessTokenTTL() time.Duration {
	return repo.accessTokenTTL
}

func (repo *authRepository) RefreshTokenTTL() time.Duration {
	return repo.refreshTokenTTL
}

func (repo *authRepository) Login(ctx context.Context, authLogin model.Auth) (err error) {
	// use defer, naked return and then wrap err to give context & reduce duplication
	defer func() {
		if err != nil {
			err = fmt.Errorf("service.authRepository.Login: %w", err)
		}
	}()
	//postgres
	var sid string
	err = repo.postgres.QueryRowContext(
		ctx, insertAuthLogin,
		authLogin.SID, authLogin.OwnerID, authLogin.Email, authLogin.RefreshToken,
		authLogin.Jti, time.Now().Add(repo.refreshTokenTTL)).Scan(&sid)

	if err != nil {
		return
	}

	// session must be created via login flow.
	key := fmt.Sprintf(sessionKeyFormat, authLogin.SID)
	err = repo.setRedisSession(ctx, key, authLogin)
	if err != nil {
		return
	}

	return nil
}

func (repo *authRepository) Session(ctx context.Context, sid string) (authLogoutResponse *model.Auth, errNoRow error, err error) {

	// get sid from cache if miss get from database
	key := fmt.Sprintf(sessionKeyFormat, sid)
	authLogoutResponse, err = repo.redisSession(ctx, key)
	if err != nil {
		return nil, nil, fmt.Errorf("service.authRepository.Session: %w", err)
	}

	// redis no entry
	if authLogoutResponse == nil || errors.Is(err, redisV8.Nil) {
		authLogoutResponse = &model.Auth{}
		// handling error cache miss
		// get from persistence database
		err = repo.postgres.QueryRowContext(ctx, getSessionBySessionID, sid).Scan(
			&authLogoutResponse.SID,
			&authLogoutResponse.OwnerID,
			&authLogoutResponse.Email,
			&authLogoutResponse.Jti,
			&authLogoutResponse.RefreshToken,
			&authLogoutResponse.ExpiredAt,
		)

		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("service.authRepository.Session: %w", err), nil
		}

		if err != nil {
			return nil, nil, fmt.Errorf("service.authRepository.Session: %w", err)
		}
		err = repo.setRedisSession(ctx, key, *authLogoutResponse)
		if err != nil {
			return nil, nil, fmt.Errorf("service.authRepository.Session: %w", err)
		}
		return authLogoutResponse, nil, nil
	}

	authLogoutResponse.SID = sid
	return authLogoutResponse, nil, nil
}

func (repo *authRepository) DeleteSession(ctx context.Context, sid string) error {
	// postgres
	var nAffected int64
	res, err := repo.postgres.ExecContext(ctx, deleteSession, sid)
	if err != nil {
		return fmt.Errorf("service.authRepository.DeleteSession: %w", err)
	}
	nAffected, err = res.RowsAffected()
	if nAffected == 0 && err == nil {
		return fmt.Errorf("service.authRepository.DeleteSession: no rows")
	}
	// redis
	keySID := fmt.Sprintf(sessionKeyFormat, sid)
	err = repo.deleteRedisSession(ctx, keySID)
	if err != nil {
		return fmt.Errorf("service.authRepository.DeleteSession: %w", err)
	}

	return nil
}

func (repo *authRepository) GetSessionIDByEmail(ctx context.Context, email string) (sessionID string, err error) {
	key := fmt.Sprintf(sessionByEmailFormat, email)
	sid, err := repo.redis.Get(ctx, key).Result()
	if errors.Is(err, redisV8.Nil) {
		return "", nil
	}
	if err != nil {
		err = fmt.Errorf("service.authRepository.GetSessionByEMail: %w", err)
		return "", err
	}

	return sid, nil

}
