package repository

import "strings"

func looksLikeMissingRelation(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "relation") && strings.Contains(s, "does not exist")
}

func looksLikeMissingColumn(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "column") && strings.Contains(s, "does not exist")
}
