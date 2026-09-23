package providersn

import "regexp"

var (
	reTalentaSNRefLine = regexp.MustCompile(`(?i)\bSN\s*/\s*Ref\s*:\s*([^\r\n]+)`)
	reMultikomSN       = regexp.MustCompile(`(?i)\bSN:\s*([^\r\n]+?)(?:\.\s*(?:Stok|Saldo)\b|$)`)
	reAJSSN            = regexp.MustCompile(`(?i)\bSN:\s*([^\r\n]+?)(?:\.\s*(?:Stok|Saldo)\b|$)`)
	reSagaraSNRefLine  = regexp.MustCompile(`(?i)\bSN\s*/\s*Ref\s*:\s*([^\r\n]+)`)
	reSagaraSN         = regexp.MustCompile(`(?i)\bSN:\s*([^\r\n]+?)(?:\s{2,}|\s+saldo\b|$)`)
	reMinionsSNRefLine = regexp.MustCompile(`(?i)\bSN\s*/\s*Ref\s*:\s*([^\r\n]+)`)
	reMinionsSN        = regexp.MustCompile(`(?i)\bSN:\s*([^\r\n]+?)(?:\s{2,}|\s+saldo\b|$)`)
	reYuscomNoResi     = regexp.MustCompile(`(?i)\bNO\s+RESI\s*:\s*([A-Z0-9\-]+)\b`)
	reYuscomExplicitSN = regexp.MustCompile(`(?i)\b(?:NO\s+REFERENSI|REFF|REF|IDT)\s*:\s*([A-Z0-9\-]+)\b`)
	reYuscomPLNToken   = regexp.MustCompile(`^\s*([0-9]{4}(?:-[0-9]{4}){4})\b`)
)
