package helloservice

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-kit/log"
)

// Service describes a service that say hello to someone.
type Service interface {
	SayHello(ctx context.Context, a string) (string, error)
}

// New returns a basic Service with all of the expected middlewares wired in.
func New(logger log.Logger) Service {
	var svc Service
	{
		svc = NewBasicService()
		svc = LoggingMiddleware(logger)(svc)
	}
	return svc
}

var (
	// ErrNameTooLong is an arbitrary business rule for the SayHello method.
	ErrNameTooLong = errors.New("name can't over than 10 bytes")
)

// NewBasicService returns a naïve, stateless implementation of Service.
func NewBasicService() Service {
	return basicService{}
}

type basicService struct{}

// SayHello impelemtns Service
func (s basicService) SayHello(_ context.Context, a string) (string, error) {
	if a != "" {
		if len(a) > 10 {
			return "", ErrNameTooLong
		}
		return fmt.Sprintf("Hello %v", a), nil
	} else {
		return "Hello World, this is a basic CICD demo.", nil
	}
}
