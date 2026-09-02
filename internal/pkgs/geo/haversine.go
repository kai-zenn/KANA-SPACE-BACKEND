package geo

import "math"

const metersPerDegreeLat = 111045.0 // ~111 km per derajat lintang

func BoundingBox(lat, lng, radiusMeters float64) (minLat, maxLat, minLng, maxLng float64) {
	latDelta := radiusMeters / metersPerDegreeLat
	lngDelta := radiusMeters / (metersPerDegreeLat * math.Cos(lat*math.Pi/180))
	
	return lat - latDelta, lat + latDelta, lng - lngDelta, lng + lngDelta
}
