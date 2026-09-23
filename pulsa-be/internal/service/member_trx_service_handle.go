package service

import (
	"context"
	"strings"

	trxmemberdto "pulsa2/internal/dto/trx_member"
	"pulsa2/internal/helper"
)

func (h *MemberTrxService) Handle(ctx context.Context, apiKey, clientIP string, in trxmemberdto.TrxRequest) (any, *ServiceError) {
	in.Commands = strings.TrimSpace(strings.ToUpper(in.Commands))
	in.Product = helper.NormalizeInternalProductCode(in.Product)
	in.Dest = strings.TrimSpace(in.Dest)
	in.RefID = strings.TrimSpace(in.RefID)
	in.PIN = strings.TrimSpace(in.PIN)

	if in.Commands == "" {
		return nil, &ServiceError{Kind: ErrBadRequest, Message: "commands required"}
	}

	switch in.Commands {
	case "PAY":
		if in.Product == "" {
			return nil, &ServiceError{Kind: ErrBadRequest, Message: "product required untuk PAY"}
		}
		if in.Dest == "" || in.RefID == "" {
			return nil, &ServiceError{Kind: ErrBadRequest, Message: "dest/refid required untuk PAY"}
		}
		if in.Qty <= 0 {
			return nil, &ServiceError{Kind: ErrBadRequest, Message: "qty harus > 0"}
		}
	case "INQ":
		if in.Product == "" {
			return nil, &ServiceError{Kind: ErrBadRequest, Message: "product required untuk INQ"}
		}
		if in.Dest == "" || in.RefID == "" {
			return nil, &ServiceError{Kind: ErrBadRequest, Message: "dest/refid required untuk INQ"}
		}
		in.Qty = 1
	case "STATUS-PAY":
		if in.Product == "" {
			return nil, &ServiceError{Kind: ErrBadRequest, Message: "product required untuk STATUS-PAY"}
		}
		if in.Dest == "" || in.RefID == "" {
			return nil, &ServiceError{Kind: ErrBadRequest, Message: "dest/refid required untuk STATUS-PAY"}
		}
		if in.Qty <= 0 {
			return nil, &ServiceError{Kind: ErrBadRequest, Message: "qty wajib nominal untuk STATUS-PAY"}
		}
	case "PRODUK":
	case "SALDO":
	case "DEPOSIT":
		if in.RefID == "" {
			return nil, &ServiceError{Kind: ErrBadRequest, Message: "refid required untuk DEPOSIT"}
		}
		if in.Qty <= 0 {
			return nil, &ServiceError{Kind: ErrBadRequest, Message: "qty harus > 0 untuk DEPOSIT"}
		}
	default:
		return nil, &ServiceError{Kind: ErrBadRequest, Message: "commands tidak didukung (hanya PAY/INQ/STATUS-PAY/PRODUK/SALDO/DEPOSIT)"}
	}

	handleTimeout := defaultHandleTimeout
	handleParent := ctx
	if in.Commands == "PAY" || in.Commands == "INQ" {
		handleTimeout = payInqHandleTimeout
		handleParent = context.Background()
	}
	ctx, cancel := context.WithTimeout(handleParent, handleTimeout)
	defer cancel()

	auth, err := h.authByAPIKeyCached(ctx, apiKey)
	if err != nil {
		return nil, &ServiceError{Kind: ErrUnauthorized, Message: "unauthorized"}
	}

	h.logf("MASUK cmd=%s produk=%s tujuan=%s qty=%d refid=%s member_id=%d ip=%s",
		in.Commands, in.Product, in.Dest, in.Qty, in.RefID, auth.MemberID, clientIP)

	allowed, err := h.isIPAllowedForMemberCached(ctx, auth.MemberID, clientIP)
	if err != nil {
		return nil, &ServiceError{Kind: ErrBadRequest, Message: err.Error()}
	}
	if !allowed {
		return nil, &ServiceError{Kind: ErrForbidden, Message: "ip not allowed"}
	}

	if in.PIN == "" || !auth.PinHash.Valid {
		return nil, &ServiceError{Kind: ErrBadRequest, Message: "pin required"}
	}
	if !h.verifyMemberPIN(auth.MemberID, auth.PinHash.String, in.PIN) {
		return nil, &ServiceError{Kind: ErrForbidden, Message: "invalid pin"}
	}
	if in.Commands == "PAY" || in.Commands == "STATUS-PAY" {
		if markerErr := h.recordSMPAYRefSource(ctx, auth, in); markerErr != nil {
			return nil, markerErr
		}
	}

	if in.Commands == "PRODUK" {
		if !helper.IsH2HRole(auth.Role) {
			return nil, &ServiceError{Kind: ErrForbidden, Message: "produk hanya untuk member h2h"}
		}
		if h.H2HProdukRepo == nil {
			return nil, &ServiceError{Kind: ErrUpstream, Message: "repo produk belum tersedia"}
		}

		rows, err := h.H2HProdukRepo.ListByMember(ctx, auth.MemberID, in.Product, "", "")
		if err != nil {
			return nil, &ServiceError{Kind: ErrUpstream, Message: "gagal mengambil daftar produk"}
		}
		return trxmemberdto.MapProdukListResponse(in.Product, rows), nil
	}

	if in.Commands == "SALDO" {
		saldo, err := h.MemberRepo.GetSaldo(ctx, auth.MemberID)
		if err != nil {
			return nil, &ServiceError{Kind: ErrUpstream, Message: "gagal mengambil saldo"}
		}
		return map[string]any{
			"ok":      true,
			"command": "SALDO",
			"balance": saldo,
		}, nil
	}

	if in.Commands == "DEPOSIT" {
		note := strings.TrimSpace(in.Berita)
		if note == "" {
			note = "deposit api"
		}
		if h.BankRepo != nil {
			if _, _, err := h.BankRepo.CreditMemberDepositAndSystemQRTPIfNeeded(ctx, auth.MemberID, in.RefID, in.Qty, note); err != nil {
				return nil, &ServiceError{Kind: ErrUpstream, Message: "gagal deposit saldo"}
			}
		} else if err := h.MemberRepo.CreditDompet(ctx, auth.MemberID, in.RefID, in.Qty, "DEPOSIT_API", note); err != nil {
			return nil, &ServiceError{Kind: ErrUpstream, Message: "gagal deposit saldo"}
		}
		saldo, err := h.MemberRepo.GetSaldo(ctx, auth.MemberID)
		if err != nil {
			return nil, &ServiceError{Kind: ErrUpstream, Message: "gagal mengambil saldo"}
		}
		return map[string]any{
			"ok":      true,
			"command": "DEPOSIT",
			"refid":   in.RefID,
			"amount":  in.Qty,
			"balance": saldo,
		}, nil
	}

	billingNominal := in.Qty
	productRuleSource := "legacy"
	chargeReceiverEligible := false
	openAmountUsesBankDest := false
	productJamBuka := ""
	productJamTutup := ""
	if in.Commands == "PAY" {
		rule, rErr := h.MemberRepo.GetProdukPricingRuleBySKU(ctx, in.Product)
		if rErr != nil {
			h.logf("PAY gagal lookup rule produk produk=%s err=%v", in.Product, rErr)
			return trxmemberdto.MapBusinessStatusResponse(in.RefID, 3, "gagal baca master produk"), nil
		}

		if rule != nil {
			productJamBuka = rule.JamBuka
			productJamTutup = rule.JamTutup
			switch rule.TipeHarga {
			case "FIXED":
				if in.Qty != 1 {
					return trxmemberdto.MapBusinessStatusResponse(in.RefID, 3, "qty untuk produk FIXED harus 1"), nil
				}
				if rule.Nominal == nil || *rule.Nominal <= 0 {
					return trxmemberdto.MapBusinessStatusResponse(in.RefID, 3, "konfigurasi produk FIXED belum valid (nominal kosong)"), nil
				}
				billingNominal = *rule.Nominal
				productRuleSource = "master_fixed"
			case "OPEN_AMOUNT":
				if in.Qty <= 0 {
					return trxmemberdto.MapBusinessStatusResponse(in.RefID, 3, "qty untuk OPEN_AMOUNT harus > 0"), nil
				}
				billingNominal = in.Qty
				productRuleSource = "master_open_amount_tipe_only"
				openAmountUsesBankDest = helper.ClassifyH2HFeeCategoryFromMetadata(in.Product, rule.KategoriNama, rule.BrandNama) == helper.H2HFeeCategoryBank
				chargeReceiverEligible = isChargeReceiverEligibleProduct(in.Product, rule.TipeHarga, rule.KategoriNama, rule.BrandNama)
			default:
				return trxmemberdto.MapBusinessStatusResponse(in.RefID, 3, "tipe_harga produk tidak didukung"), nil
			}
		} else {
			return trxmemberdto.MapBusinessStatusResponse(in.RefID, 3, "produk tidak ditemukan"), nil
		}
	}

	// Validasi format tujuan untuk nominal bebas (OPEN_AMOUNT)
	// Bank transfer pakai nomor rekening — hanya validasi angka, tidak harus 08/628
	if productRuleSource == "master_open_amount_tipe_only" && in.Commands == "PAY" {
		if openAmountUsesBankDest {
			if !helper.IsValidBankAccountFormat(in.Dest) {
				return trxmemberdto.MapBusinessStatusResponse(in.RefID, 3, "nomor rekening tidak valid, harus angka"), nil
			}
		} else {
			if !helper.IsValidDestinationFormat(in.Dest) {
				return trxmemberdto.MapBusinessStatusResponse(in.RefID, 3, "format tujuan salah, harus diawali 08 atau 628"), nil
			}
		}
	}

	if statusResp, handled := h.handleStatusPayBranch(ctx, auth, in); handled {
		if statusResp.Err != nil {
			return trxmemberdto.MapBusinessStatusResponse(in.RefID, 3, statusResp.Err.Message), nil
		}
		return statusResp.Body, nil
	}

	out := h.handlePayInqBranch(ctx, auth, in, billingNominal, productRuleSource, chargeReceiverEligible, productJamBuka, productJamTutup)
	if out.Err != nil {
		return trxmemberdto.MapBusinessStatusResponse(in.RefID, 3, out.Err.Message), nil
	}
	return out.Body, nil
}
