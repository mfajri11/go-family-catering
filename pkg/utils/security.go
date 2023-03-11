package utils

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

// var (
// 	hashPassword     func([]byte, int) ([]byte, error) = bcrypt.GenerateFromPassword
// 	validatePassword func([]byte, []byte) error        = bcrypt.CompareHashAndPassword
// )

// func resetHashPassword() {
// 	hashPassword = bcrypt.GenerateFromPassword
// }

// func resetValidatePassword() {
// 	validatePassword = bcrypt.CompareHashAndPassword
// }

var (
	HashPassword     func(password string) (string, error)
	ValidatePassword func(string, string) error
	// GenerateRandomInt64  func() (int64, error)
	// GenerateAccessToken  func(time.Duration, string) (string, string, error)
	// GenerateRefreshToken func(string, time.Duration) (string, error)
	// GenerateRandomString func(int) (string, error)
	GenerateToken func(expire time.Duration, jti string, email string) (string, error)
	ValidateToken func(token string) (*JwtClaims, error)

	secretKeyAccessToken  string = "secretKeyAccessToken"  // TODO: refactored, must not be hardcoded
	secretKeyRefreshToken string = "secretKeyRefreshToken" // TODO: refactored, must not be hardcoded
)

// const letters string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

const (
	accessTokenType  string = "at"
	refreshTokenType string = "rt"
)

type JwtClaims struct {
	Type             string `json:"type,omitempty"`
	Email            string `json:"email,omitempty"`
	ForResetPassword bool   `json:"for_reset_password,omitempty"`
	forTesting       bool   // use only for bypassing
	jwt.StandardClaims
}

func (j *JwtClaims) IsForRefreshToken() bool {
	return j.forTesting || (j.Type == refreshTokenType)
}

func (j *JwtClaims) IsForResetPassword() bool {
	return j.forTesting || j.ForResetPassword
}

// func (j *JwtClaims) Email() string {
// 	return j.email
// }

func NewJWTClaimTesting(id string) *JwtClaims {
	j := JwtClaims{}
	j.forTesting = true
	j.Id = id
	return &j
}

func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(hashed), nil
}

func validatePassword(password, candidate string) error {
	decodeCandidate, err := base64.StdEncoding.DecodeString(candidate)
	if err != nil {
		return err
	}
	return bcrypt.CompareHashAndPassword(decodeCandidate, []byte(password))
}

const maxUint64 = ^int64(0)

// func generateRandomInt64() (int64, error) {
// 	nInt64, err := rand.Int(rand.Reader, big.NewInt(maxUint64))
// 	if err != nil {
// 		return 0, err
// 	}
// 	return nInt64.Int64(), nil
// }

// func generateAccessToken(expire time.Duration, email string) (string, string, error) {

// 	createdAt := time.Now()
// 	claims := JwtClaims{
// 		_type: accessTokenType,
// 		StandardClaims: jwt.StandardClaims{
// 			IssuedAt:  createdAt.Unix(),
// 			ExpiresAt: createdAt.Add(expire).Unix(),
// 			Issuer:    "family-catering.com", // TODO: do not hardcode, retrieve from config
// 		},
// 	}

// 	id, err := generateRandomString(32)
// 	if err != nil {
// 		return "", "", err
// 	}

// 	if email != "" {
// 		claims.email = email
// 		claims.ForResetPassword = true
// 		claims.Id = id
// 	}

// 	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
// 	tokenStr, err := token.SignedString(secretKeyAccessToken) // TODO: refactored use secret keys from config
// 	if err != nil {
// 		return "", "", err
// 	}
// 	return tokenStr, id, nil
// }

func generateToken(expire time.Duration, id, email string) (string, error) {
	createdAt := time.Now()
	claims := JwtClaims{
		Type: accessTokenType,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  createdAt.Unix(),
			ExpiresAt: createdAt.Add(expire).Unix(),
		},
	}
	var secret []byte
	secret = []byte(secretKeyAccessToken)
	if id != "" {
		claims.Id = id

		if email != "" {
			claims.Email = email
			claims.ForResetPassword = true
			claims.Type = accessTokenType

		} else {
			claims.Type = refreshTokenType
			secret = []byte(secretKeyRefreshToken)
		}
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)

}

// func generateRefreshToken(jti string, expire time.Duration) (string, error) {
// 	createdAt := time.Now()
// 	claims := JwtClaims{
// 		_type: refreshTokenType,
// 		StandardClaims: jwt.StandardClaims{
// 			IssuedAt:  createdAt.Unix(),
// 			ExpiresAt: createdAt.Add(expire).Unix(),
// 			Issuer:    "Family-Catering", // TODO: do not hardcode, retrieve from config
// 			Id:        jti,
// 		},
// 	}

// 	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
// 	return token.SignedString(secretKeyRefreshToken)

// }

func validateToken(tokenString string) (*JwtClaims, error) {
	claims := &JwtClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		var secret string
		// validate method
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("utils.ValidateToken: unexpected signing method, want HS256 got %s", t.Header["alg"])
		}

		switch claims.Type {
		case accessTokenType:
			secret = secretKeyAccessToken
		case refreshTokenType:
			secret = secretKeyRefreshToken
		default:
			secret = ""
		}

		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("utils.ValidateToken: err %w", err)
	}
	payload, ok := token.Claims.(*JwtClaims)
	if !ok {
		return nil, fmt.Errorf("utils.ValidateToken: invalid payload data")
	}

	return payload, nil
}

// func generateRandomString(n int) (string, error) {
// 	var err error
// 	b := strings.Builder{}
// 	maxRange := len(letters)
// 	for i := 0; i < n; i++ {
// 		err = b.WriteByte(letters[mRand.Intn(maxRange)])
// 		if err != nil {
// 			return "", err
// 		}
// 	}

// 	return b.String(), nil
// }
