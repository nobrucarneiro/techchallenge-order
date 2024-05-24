package usecases

import (
	"g37-lanchonete/internal/core/usecases/dto"
	"g37-lanchonete/internal/infra/drivers/auth"

	log "github.com/sirupsen/logrus"
)

type AuthorizerUsecase interface {
	AuthorizeUser(cpf string) (dto.AuthorizerResponse, error)
}

type authorizerUsecase struct {
	authorizer auth.Authorizer
}

func NewAuthorizerUsecase(authorizer auth.Authorizer) AuthorizerUsecase {
	return authorizerUsecase{
		authorizer: authorizer,
	}
}

func (u authorizerUsecase) AuthorizeUser(cpf string) (dto.AuthorizerResponse, error) {
	authorizerResponse, err := u.authorizer.AuthorizeUser(cpf)
	if err != nil {
		log.Errorf("failed to authorize user, error: %v", err)
		return dto.AuthorizerResponse{}, err
	}

	return authorizerResponse, nil
}