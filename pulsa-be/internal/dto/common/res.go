package commondto

type ErrorResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

type OKResponse struct {
	Ok bool `json:"ok"`
}

type IDResponse struct {
	Ok bool  `json:"ok"`
	ID int64 `json:"id"`
}

type ItemResponse struct {
	Ok   bool `json:"ok"`
	Item any  `json:"item"`
}

type ListResponse struct {
	Ok    bool `json:"ok"`
	Items any  `json:"items"`
}
