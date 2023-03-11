package utils

import (
	"context"
)

var (
	keys             map[string]*contextKey = make(map[string]*contextKey)
	ValueContext     func(context.Context, string) interface{}
	ContextWithValue func(context.Context, string, interface{}) context.Context
)

type contextKey struct {
	key string
}

func (c contextKey) String() string {
	return c.key
}

func getOrNewContextKey(key string) *contextKey {
	val, ok := keys[key]
	if !ok {
		ctxKey := contextKey{key}
		keys[key] = &ctxKey
		val = &ctxKey
	}

	return val
}

func valueContext(ctx context.Context, key string) interface{} {
	ctxKey := getOrNewContextKey(key)
	// ctxKey := contextKey{key}
	return ctx.Value(ctxKey)
}

func contextWithValue(ctx context.Context, key string, val interface{}) context.Context {
	ctxKey := getOrNewContextKey(key)
	// ctxKey := contextKey{key}
	return context.WithValue(ctx, ctxKey, val)
}
