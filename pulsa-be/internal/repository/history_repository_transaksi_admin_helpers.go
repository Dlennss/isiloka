package repository

import "strings"

func expandKodeProdukFilters(kodeProduk string) []string {
	k := strings.ToUpper(strings.TrimSpace(kodeProduk))
	if k == "" {
		return nil
	}
	switch k {
	case "DANA":
		return []string{"DANA", "DNID"}
	default:
		return []string{k}
	}
}
