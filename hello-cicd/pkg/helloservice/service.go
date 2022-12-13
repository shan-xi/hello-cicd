package helloservice

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-kit/kit/log"
)

type Service interface {
	SayHello(ctx context.Context, a string) (string, error)
}

func New(logger log.Logger) Service {
	var svc Service
	{
		svc = NewBasicService()
		svc = LoggingMiddleware(logger)(svc)
	}
	return svc
}

var (
	ErrNameTooLong = errors.New("name can't over than 10 bytes")
)

func NewBasicService() Service {
	return basicService{}
}

type basicService struct{}

func (s basicService) SayHello(_ context.Context, a string) (string, error) {
	if a != "" {
		if len(a) > 10 {
			return "", ErrNameTooLong
		}
		return fmt.Sprintf("Hello %v", a), nil
	} else {
		return fmt.Sprint("Hello World"), nil
	}
}
