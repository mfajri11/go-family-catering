package utils

import "family-catering/pkg/consts"

// any function should be pure or method which not remember it internal state
// this is due to code design on this package in order to make mock/patch easier

func init() {
	HashPassword = hashPassword
	ValidatePassword = validatePassword
	// GenerateRandomInt64 = generateRandomInt64
	// GenerateAccessToken = generateAccessToken
	// GenerateRefreshToken = generateRefreshToken
	GenerateToken = generateToken
	ValidateRequest = validateRequest
	ValidateToken = validateToken
	ValueContext = valueContext
	ContextWithValue = contextWithValue
	keys[consts.CtxKeyAuthorization] = &contextKey{consts.CtxKeyAuthorization}
	keys[consts.CtxKeySID] = &contextKey{consts.CtxKeySID}
	// GenerateRandomString = generateRandomString
}
