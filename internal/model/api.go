package model

type CreateIDReq struct {
	URL string `json:"url" validate:"required,url"`
}

type CreateIDResp struct {
	Result string `json:"result"`
}
