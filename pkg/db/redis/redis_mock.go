package redis

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/elliotchance/redismock/v8"
)

func NewMockRedisClient() *redismock.ClientMock {
	return redismock.NewMock()

}

func NewMockWithMiniRedisClient(t *testing.T) (*redismock.ClientMock, error) {
	s := miniredis.RunT(t)
	c, err := New(s.Addr(), "")
	if err != nil {
		return nil, err
	}
	niceMock := redismock.NewNiceMock(c)

	return niceMock, nil
}
