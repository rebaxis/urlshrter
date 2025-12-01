package model

type (
	CreateIDReq struct {
		URL string `json:"url" valid:"required,url"`
	}

	CreateIDResp struct {
		Result string `json:"result" valid:"required"`
	}

	CreateIDBatchReq struct {
		Batch  []BatchEntReq
		UserID string `json:"user_id" valid:"required"`
	}

	BatchEntReq struct {
		CorrelationID string `json:"correlation_id" valid:"required"`
		OriginalURL   string `json:"original_url" valid:"required,url"`
	}

	CreateIDBatchResp struct {
		Batch []BatchEntResp
	}

	BatchEntResp struct {
		CorrelationID string `json:"correlation_id" valid:"required"`
		ShortURL      string `json:"short_url" valid:"required"`
	}

	BatchEntUserResp struct {
		OriginalURL string `json:"original_url" valid:"required,url"`
		ShortURL    string `json:"short_url" valid:"required"`
	}

	GetBatchEntUserResp struct {
		Batch []BatchEntUserResp
	}
)
