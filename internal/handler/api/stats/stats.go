// Package stats содержит обработчик эндпоинта статистики сервиса.
package stats

import (
	"encoding/json"
	"net/http"

	"github.com/rebaxis/urlshrter/internal/config/logger"
	"github.com/rebaxis/urlshrter/internal/config/shortener"
	"github.com/rebaxis/urlshrter/internal/lib"
	"github.com/rebaxis/urlshrter/internal/model"
	"github.com/rebaxis/urlshrter/internal/service"
)

// GetStats возвращает HTTP-обработчик для GET /api/internal/stats.
// Доступ разрешён только клиентам, чей IP-адрес (из заголовка X-Real-IP)
// входит в доверенную подсеть opts.TrustedSubnet.
// Если TrustedSubnet не задан или IP не входит в подсеть — возвращает 403 Forbidden.
func GetStats(svc service.URLService, opts shortener.Opts) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !lib.IsIPTrusted(r.Header.Get("X-Real-IP"), opts.TrustedSubnet) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		urlCount, userCount, err := svc.GetStats()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		resp := model.StatsResp{
			URLs:  urlCount,
			Users: userCount,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		log := logger.GetLogger()
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Log.Errorln("failed to encode stats response", err)
		}
	}
}
