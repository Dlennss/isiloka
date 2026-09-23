package service

import (
	"context"
	"strings"

	"pulsa2/internal/repository"
)

type AppAdService struct {
	repo *repository.AppAdRepository
}

func NewAppAdService(repo *repository.AppAdRepository) *AppAdService {
	return &AppAdService{repo: repo}
}

func (s *AppAdService) List(ctx context.Context, q string) ([]repository.AppAdRow, error) {
	return s.repo.List(ctx, q)
}

func (s *AppAdService) ListActive(ctx context.Context) ([]repository.AppAdRow, error) {
	return s.repo.ListActive(ctx)
}

func (s *AppAdService) Get(ctx context.Context, id int64) (*repository.AppAdRow, error) {
	return s.repo.Get(ctx, id)
}

func (s *AppAdService) Create(ctx context.Context, in repository.AppAdUpsertInput) (int64, error) {
	in.Judul = strings.TrimSpace(in.Judul)
	in.Keterangan = strings.TrimSpace(in.Keterangan)
	in.ImageURL = strings.TrimSpace(in.ImageURL)
	in.LinkURL = strings.TrimSpace(in.LinkURL)
	if in.ImageURL == "" {
		return 0, errBadRequest("image_url wajib diisi")
	}
	if in.Judul == "" && in.Keterangan == "" {
		return 0, errBadRequest("judul atau keterangan wajib diisi")
	}
	if in.Urutan < 0 {
		return 0, errBadRequest("urutan tidak valid")
	}
	return s.repo.Create(ctx, in)
}

func (s *AppAdService) Update(ctx context.Context, in repository.AppAdUpsertInput) error {
	in.Judul = strings.TrimSpace(in.Judul)
	in.Keterangan = strings.TrimSpace(in.Keterangan)
	in.ImageURL = strings.TrimSpace(in.ImageURL)
	in.LinkURL = strings.TrimSpace(in.LinkURL)
	if in.ImageURL == "" {
		return errBadRequest("image_url wajib diisi")
	}
	if in.Judul == "" && in.Keterangan == "" {
		return errBadRequest("judul atau keterangan wajib diisi")
	}
	if in.Urutan < 0 {
		return errBadRequest("urutan tidak valid")
	}
	return s.repo.Update(ctx, in)
}

func (s *AppAdService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

type badRequestError struct {
	msg string
}

func (e badRequestError) Error() string { return e.msg }

func errBadRequest(msg string) error {
	return badRequestError{msg: msg}
}
