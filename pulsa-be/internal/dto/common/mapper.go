package commondto

func MapError(msg string) ErrorResponse {
	return ErrorResponse{Ok: false, Error: msg}
}

func MapOK() OKResponse {
	return OKResponse{Ok: true}
}

func MapID(id int64) IDResponse {
	return IDResponse{Ok: true, ID: id}
}

func MapItem(item any) ItemResponse {
	return ItemResponse{Ok: true, Item: item}
}

func MapList(items any) ListResponse {
	return ListResponse{Ok: true, Items: items}
}
