// Package ext_geojson provides a simple GeoJSON matcher that allows you to
// match against message values that are GeoJSON and fall within a polygon
// defined in the subscription. Subscription values are polygons and message
// values are points or polygons.
// More GeoJSON matchers can be added to this package for things like intersection,
// distance etc.
package ext_geojson

import _ "strings"
import "github.com/paulsmith/gogeos/geos"
import "log"
import "errors"
import _ "fmt"

// The array in the JSON is interface{}, this helper casts and converts
// to a geos Coord value.
// TODO: Needs error handling
func toGeosCoords(s []interface{}) []geos.Coord {
	var geomcoords []geos.Coord
	for _, cp := range s {
		// cp is the untyped coordinate pair
		tcp := cp.([]interface{})
		x := tcp[0].(float64)
		y := tcp[1].(float64)
		geomcoords = append(geomcoords, geos.NewCoord(x, y))
	}

	return geomcoords
}

// Convert the GeoJSON map to a geometry object to work with
func JsonToGeometry(g map[string]interface{}) (*geos.Geometry, error) {
	var geomcoords []geos.Coord
	var geom *geos.Geometry

	geotype, ok := g["type"].(string)
	log.Println("T", geotype)
	if !ok {
		return nil, errors.New("ext_geojson: Missing GeoJSON type")
	}
	if geotype != "Polygon" && geotype != "Point" {
		return nil, errors.New("ext_geojson: Invalid GeoJSON type. Point and Polygon only supported.")
	}

	// Build GEOS coordinate array
	coords := g["coordinates"].([]interface{})
	if len(coords) == 0 {
		return nil, errors.New("Can't handle GeoJSON Polygons/Points with no coordinates")
	}

	if geotype == "Polygon" {
		if len(coords) > 1 {
			return nil, errors.New("Can't handle GeoJSON Polygons with holes")
		}

		// coords[0] is the outer ring as per GeoJSON spec
		geomcoords = toGeosCoords(coords[0].([]interface{}))
		geom, _ = geos.NewPolygon(geomcoords)
		log.Println("Polygon:", geom)
	} else {
		// Point type
		x := coords[0].(float64)
		y := coords[1].(float64)
		geom, _ = geos.NewPoint(geos.NewCoord(x, y))
		log.Println("Point:", geom)
	}

	return geom, nil
}

// ExtGeoJSONWithin matches against a message value that is a Point or Polygon
// and is a positive match when the message value is 'within' the subscription
// polygon completely.
func ExtGeoJSONWithin(mval interface{}, sval map[string]interface{}) bool {
	var messageGeom *geos.Geometry
	var specGeom *geos.Geometry

	sjson, ok := sval["geojson"].(map[string]interface{})
	if !ok {
		return false
	}

	mjson := mval.(map[string]interface{})

	if sgeom, err := JsonToGeometry(sjson); err != nil {
		log.Println("Error building spec GeoJSON:", err)
		return false
	} else {
		specGeom = sgeom
	}
	if mgeom, err := JsonToGeometry(mjson); err != nil {
		log.Println("Error building message GeoJSON:", err)
		return false
	} else {
		messageGeom = mgeom
	}

	if isContained, err := messageGeom.Within(specGeom); err != nil {
		log.Println("Error in Geom.contains:", err)
	} else {
		if isContained {
			log.Println("Geom.contains TRUE")
			return true
		}
	}

	log.Println("Geom.contains FALSE")
	return false

}
