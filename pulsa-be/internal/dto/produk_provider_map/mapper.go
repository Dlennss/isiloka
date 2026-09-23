package produkprovidermapdto

import commondto "pulsa2/internal/dto/common"

func MapError(msg string) ErrorResponse {
	return commondto.MapError(msg)
}

func MapID(id int64) IDResponse {
	return commondto.MapID(id)
}

func MapOK() OKResponse {
	return commondto.MapOK()
}

func MapItem(item any) ItemResponse {
	return commondto.MapItem(item)
}

func MapList(items any) ListResponse {
	return commondto.MapList(items)
}
