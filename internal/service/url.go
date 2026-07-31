package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/keyboy21/url-shortner/internal/apperror"
	"github.com/keyboy21/url-shortner/internal/model"
	"github.com/keyboy21/url-shortner/internal/store"
)

const generatedAliasAttempts = 5

var aliasPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{3,64}$`)

type UrlService struct {
	repository store.UrlInterface
}

func NewUrlService(repository store.UrlInterface) *UrlService {
	return &UrlService{repository: repository}
}

func (s *UrlService) Create(ctx context.Context, rawUrl, alias string) (*model.Url, error) {
	rawUrl = strings.TrimSpace(rawUrl)
	alias = strings.TrimSpace(alias)

	if !isValidUrl(rawUrl) {
		return nil, apperror.ErrInvalidUrl
	}

	if alias != "" {
		if !aliasPattern.MatchString(alias) {
			return nil, apperror.ErrInvalidAlias
		}
		return s.save(ctx, rawUrl, alias)
	}

	for range generatedAliasAttempts {
		generatedAlias, err := newAlias()
		if err != nil {
			return nil, fmt.Errorf("generate alias: %w", err)
		}

		created, err := s.save(ctx, rawUrl, generatedAlias)
		if errors.Is(err, apperror.ErrAliasAlreadyUsed) {
			continue
		}
		return created, err
	}

	return nil, apperror.ErrAliasAlreadyUsed
}

func (s *UrlService) FindByAlias(ctx context.Context, alias string) (*model.Url, error) {
	alias = strings.TrimSpace(alias)
	if !aliasPattern.MatchString(alias) {
		return nil, apperror.ErrInvalidAlias
	}

	return s.repository.FindByAlias(ctx, alias)
}

func (s *UrlService) DeleteByAlias(ctx context.Context, alias string) error {
	alias = strings.TrimSpace(alias)
	if !aliasPattern.MatchString(alias) {
		return apperror.ErrInvalidAlias
	}

	if _, err := s.repository.DeleteByAlias(ctx, alias); err != nil {
		return err
	}
	return nil
}

func (s *UrlService) save(ctx context.Context, rawUrl, alias string) (*model.Url, error) {
	return s.repository.SaveUrl(ctx, &model.Url{Url: rawUrl, Alias: alias})
}

func isValidUrl(rawUrl string) bool {
	parsed, err := url.ParseRequestURI(rawUrl)
	if err != nil || parsed.Host == "" {
		return false
	}

	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func newAlias() (string, error) {
	random := make([]byte, 6)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(random), nil
}
