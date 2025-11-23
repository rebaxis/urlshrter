package model

type (
	CreateIDReq struct {
		URL string `json:"url" valid:"required,url"`
	}

	CreateIDResp struct {
		Result string `json:"result"`
	}

	CreateIDBatchReq struct {
		Batch []BatchEntReq
	}

	BatchEntReq struct {
		CorrelationId string `json:"correlation_id" valid:"required"`
		OriginalURL   string `json:"original_url" valid:"required,url"`
	}

	CreateIDBatchResp struct {
		Batch []BatchEntResp
	}

	BatchEntResp struct {
		CorrelationId string `json:"correlation_id" valid:"required"`
		ShortURL      string `json:"short_url" valid:"required"`
	}
)
