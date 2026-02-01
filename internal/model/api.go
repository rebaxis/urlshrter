// Package model содержит API модели для HTTP запросов и ответов.
package model

type (
	// CreateIDReq представляет запрос на создание короткого URL.
	// Принимается в теле POST запроса в формате JSON.
	CreateIDReq struct {
		URL string `json:"url" valid:"required,url"` // Оригинальный URL для сокращения
	}

	// CreateIDResp представляет ответ с созданным коротким URL.
	// Возвращается клиенту в формате JSON.
	CreateIDResp struct {
		Result string `json:"result" valid:"required"` // Полный короткий URL (с base URL)
	}

	// CreateIDBatchReq представляет запрос на пакетное создание коротких URL.
	// Позволяет создать множество URL за один запрос.
	CreateIDBatchReq struct {
		Batch  []BatchEntReq // Слайс URL для создания
		UserID string        `json:"user_id" valid:"required"` // Идентификатор пользователя
	}

	// BatchEntReq представляет один элемент в пакетном запросе.
	// Содержит correlation ID для связи запроса с ответом и URL для сокращения.
	BatchEntReq struct {
		CorrelationID string `json:"correlation_id" valid:"required"`   // ID для связи запроса и ответа
		OriginalURL   string `json:"original_url" valid:"required,url"` // Оригинальный URL
	}

	// CreateIDBatchResp представляет ответ на пакетное создание URL.
	// Содержит слайс созданных коротких URL с correlation ID.
	CreateIDBatchResp struct {
		Batch []BatchEntResp // Слайс результатов создания
	}

	// BatchEntResp представляет один элемент в пакетном ответе.
	// Связывает correlation ID запроса с созданным коротким URL.
	BatchEntResp struct {
		CorrelationID string `json:"correlation_id" valid:"required"` // ID из запроса
		ShortURL      string `json:"short_url" valid:"required"`      // Созданный короткий URL
	}

	// BatchEntUserResp представляет один URL пользователя в ответе.
	// Используется для получения списка URL конкретного пользователя.
	BatchEntUserResp struct {
		OriginalURL string `json:"original_url" valid:"required,url"` // Оригинальный URL
		ShortURL    string `json:"short_url" valid:"required"`        // Короткий URL
	}

	// GetBatchEntUserResp представляет ответ со списком URL пользователя.
	// Возвращается при запросе всех URL конкретного пользователя.
	GetBatchEntUserResp struct {
		Batch []BatchEntUserResp // Слайс URL пользователя
	}

	// BatchDeleteEntByUserReq представляет запрос на удаление нескольких URL.
	// Содержит список коротких URL для удаления пользователем.
	BatchDeleteEntByUserReq struct {
		ShortURLs []string `valid:"required"` // Список коротких URL для удаления
	}
)
