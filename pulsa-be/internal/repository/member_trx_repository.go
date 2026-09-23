package repository

import "database/sql"

type MemberTrxRepository struct {
	MemberRepo     *MemberTrxMemberRepository
	ProviderTrx    *MemberTrxProviderTrxRepository
	ProviderWallet *MemberTrxProviderWalletRepository
	H2HProduk      *H2HProdukRepository
}

func NewMemberTrxRepository(db *sql.DB) *MemberTrxRepository {
	return &MemberTrxRepository{
		MemberRepo:     NewMemberTrxMemberRepository(db),
		ProviderTrx:    NewMemberTrxProviderTrxRepository(db),
		ProviderWallet: NewMemberTrxProviderWalletRepository(db),
		H2HProduk:      NewH2HProdukRepository(db),
	}
}
