package apperror

import "errors"

var (
	ErrInvalidUrl       = errors.New("invalid URL")
	ErrInvalidAlias     = errors.New("invalid alias")
	ErrUrlNotFound      = errors.New("URL not found")
	ErrUrlAlreadyExists = errors.New("URL already exists")
	ErrAliasAlreadyUsed = errors.New("alias already used")
)
