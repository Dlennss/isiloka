package repository

import (
	"context"

	"pulsa2/db"
	"pulsa2/model"
)

type JavapayInternalRepository struct {
	repo *db.JavapayRepo
}

func NewJavapayInternalRepository(repo *db.JavapayRepo) *JavapayInternalRepository {
	return &JavapayInternalRepository{repo: repo}
}

func (r *JavapayInternalRepository) GetLatestByRefIDAndPerintah(ctx context.Context, refID, perintah string) (*model.JavapayTrxRow, error) {
	return r.repo.GetLatestByRefIDAndPerintah(ctx, refID, perintah)
}

func (r *JavapayInternalRepository) Create(ctx context.Context, in model.JavapayTrxCreateIn, requestMentah any) (*model.JavapayTrxRow, error) {
	return r.repo.Create(ctx, in, requestMentah)
}

func (r *JavapayInternalRepository) UpdateRequestMentah(ctx context.Context, id int64, req any) error {
	return r.repo.UpdateRequestMentah(ctx, id, req)
}

func (r *JavapayInternalRepository) UpdateResult(ctx context.Context, id int64, u db.UpdateResult) error {
	return r.repo.UpdateResult(ctx, id, u)
}
