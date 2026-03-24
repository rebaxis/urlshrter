// Package lib содержит вспомогательные утилиты общего назначения.
// Предоставляет функции для генерации случайных строк и другие полезные инструменты.
package lib

import (
	"crypto/rand"
	"math/big"
	"net"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var alphabetLen = big.NewInt(int64(len(alphabet)))

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

// GenerateRandomAlphabetString генерирует случайную строку заданной длины.
// Полученная строка содержит только символы латинского алфавита.
// Возвращает ошибку, если не удалось сгенерировать случайное число.
func GenerateRandomAlphabetString(length int) (string, error) {
	b := make([]byte, length)
	for i := range b {
		num, err := rand.Int(rand.Reader, alphabetLen)
		if err != nil {
			return "", err
		}
		b[i] = alphabet[num.Int64()]
	}
	return string(b), nil
}
