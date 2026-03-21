// Package stats содержит обработчик эндпоинта статистики сервиса.
package stats

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/rebaxis/urlshrter/internal/config/shortener"
	"github.com/rebaxis/urlshrter/internal/model"
	"github.com/rebaxis/urlshrter/internal/service"
)

// GetStats возвращает HTTP-обработчик для GET /api/internal/stats.
// Доступ разрешён только клиентам, чей IP-адрес (из заголовка X-Real-IP)
// входит в доверенную подсеть opts.TrustedSubnet.
// Если TrustedSubnet не задан или IP не входит в подсеть — возвращает 403 Forbidden.
func GetStats(svc service.URLService, opts shortener.Opts) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsIPTrusted(r.Header.Get("X-Real-IP"), opts.TrustedSubnet) {
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
		json.NewEncoder(w).Encode(resp)
	}
}

// IsIPTrusted проверяет, входит ли переданный IP-адрес в указанную CIDR-подсеть.
// Возвращает false если cidr или ipStr пусты, либо содержат некорректные значения.
func IsIPTrusted(ipStr, cidr string) bool {
	if cidr == "" || ipStr == "" {
		return false
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	_, subnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	return subnet.Contains(ip)
}
