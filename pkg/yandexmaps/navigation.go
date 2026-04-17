// Package yandexmaps формирует ссылки на яндекс.Карты для построения маршрута.
// Формат rtext: широта,долгота — см. https://yandex.ru/support/maps/concept/URLs/ru
package yandexmaps

import (
	"net/url"
	"strconv"
)

const baseMaps = "https://yandex.ru/maps/"

// RouteToPointURL маршрут до точки назначения.
// Если fromLat/fromLng заданы — маршрут «от точки А к точке Б», иначе «с текущего местоположения до объекта» (~lat,lon).
func RouteToPointURL(destLat, destLon float64, fromLat, fromLon *float64) string {
	q := url.Values{}
	fl := formatCoord(destLat)
	fn := formatCoord(destLon)
	var rtext string
	if fromLat != nil && fromLon != nil && (*fromLat != 0 || *fromLon != 0) {
		rtext = formatCoord(*fromLat) + "," + formatCoord(*fromLon) + "~" + fl + "," + fn
	} else {
		rtext = "~" + fl + "," + fn
	}
	q.Set("rtext", rtext)
	q.Set("rtt", "auto")
	return baseMaps + "?" + q.Encode()
}

// SearchByAddressURL открывает карту с поиском по адресу (если координат объекта нет в БД).
func SearchByAddressURL(address string) string {
	q := url.Values{}
	q.Set("text", address)
	return baseMaps + "?" + q.Encode()
}

func formatCoord(v float64) string {
	return strconv.FormatFloat(v, 'f', 7, 64)
}
