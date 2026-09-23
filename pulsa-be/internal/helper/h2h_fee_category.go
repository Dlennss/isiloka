package helper

import "strings"

const (
	H2HFeeCategoryDana      = "DANA"
	H2HFeeCategoryGopay     = "GOPAY"
	H2HFeeCategoryOvo       = "OVO"
	H2HFeeCategoryLinkAja   = "LINKAJA"
	H2HFeeCategoryShopeePay = "SHOPEEPAY"
	H2HFeeCategoryBank      = "BANK"
	H2HFeeCategoryLainnya   = "LAINNYA"
)

func H2HFeeCategoryNames() []string {
	return []string{
		H2HFeeCategoryDana,
		H2HFeeCategoryGopay,
		H2HFeeCategoryOvo,
		H2HFeeCategoryLinkAja,
		H2HFeeCategoryShopeePay,
		H2HFeeCategoryBank,
		H2HFeeCategoryLainnya,
	}
}

func NormalizeH2HFeeCategory(code string) (string, bool) {
	normalized := strings.ToUpper(strings.TrimSpace(code))
	for _, allowed := range H2HFeeCategoryNames() {
		if normalized == allowed {
			return allowed, true
		}
	}
	return "", false
}

func H2HFeeCategoryLabel(code string) string {
	if normalized, ok := NormalizeH2HFeeCategory(code); ok {
		return normalized
	}
	return strings.ToUpper(strings.TrimSpace(code))
}

func IsBankTransferCategoryName(categoryName string) bool {
	c := strings.ToUpper(strings.TrimSpace(categoryName))
	return c != "" && strings.Contains(c, "BANK")
}

func ClassifyH2HFeeCategoryFromMetadata(productCode string, categoryName string, brandName string) string {
	if IsBankTransferCategoryName(categoryName) {
		return H2HFeeCategoryBank
	}
	if category := classifyH2HWalletAlias(brandName); category != "" {
		return category
	}
	if category := classifyH2HWalletAlias(productCode); category != "" {
		return category
	}
	return ClassifyH2HFeeCategory(productCode)
}

func classifyH2HWalletAlias(code string) string {
	p := strings.ToUpper(strings.TrimSpace(code))
	p = strings.NewReplacer(" ", "", "-", "", "_", "").Replace(p)
	switch {
	case p == "DANA":
		return H2HFeeCategoryDana
	case p == "GOPAY" || p == "GOJEK" || p == "GPAY":
		return H2HFeeCategoryGopay
	case p == "OVO":
		return H2HFeeCategoryOvo
	case p == "LINKAJA" || p == "LAJA":
		return H2HFeeCategoryLinkAja
	case p == "SHOPEE" || p == "SHOPEEPAY" || p == "SHPAY":
		return H2HFeeCategoryShopeePay
	default:
		return ""
	}
}

func ClassifyH2HFeeCategory(productCode string) string {
	p := strings.ToUpper(strings.TrimSpace(productCode))
	switch {
	case p == "DANA":
		return H2HFeeCategoryDana
	case p == "GOPAY" || p == "GOJEK" || p == "GPAY":
		return H2HFeeCategoryGopay
	case p == "OVO":
		return H2HFeeCategoryOvo
	case p == "LINKAJA" || p == "LAJA":
		return H2HFeeCategoryLinkAja
	case p == "SHOPEE" || p == "SHOPEEPAY" || p == "SHPAY":
		return H2HFeeCategoryShopeePay
	case p == "BCA" || p == "BRI" || p == "BNI" || p == "MANDIRI" || p == "BSI" ||
		p == "CIMB" || p == "DANAMON" || p == "PERMATA" || p == "BTN" || p == "OCBC" ||
		p == "MAYBANK" || p == "PANIN" || p == "MEGA" || p == "BUKOPIN" || p == "SINARMAS" ||
		p == "COMMONWEALTH" || p == "UOB" || p == "HSBC" || p == "BTPN" || p == "MUAMALAT" ||
		p == "JAGO" || p == "NEOCOMMERCE" || p == "SEABANK" || p == "ALLOBANK" || p == "BJB" ||
		strings.HasPrefix(p, "BANK"):
		return H2HFeeCategoryBank
	default:
		return H2HFeeCategoryLainnya
	}
}
