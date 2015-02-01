package ext_geojson

import _ "strings"
import "github.com/paulsmith/gogeos/geos"
import "log"
import "errors"
import _ "fmt"

/*
 * Takes a GeoJSON object `mval` and uses the GeoJSON object in `sval`
 * to determine if `mval` is within `sval` in some sense. The actual
 * geometry for `sval` is in sval['geojson']
 *
 * This extension expects sval['geojson'] to be a Polygon
 * and `sval` to be a polygon or point.
 *
 * Example spec:
 *
 * {
 * 		"_match": "geojson-within",
 		"geojson": {
 			"type": "Polygon",
 			"coordinates": [ [1,1], [0,1], [0,0], [1,0] ]
 		}
   }
*/

/*
 * Takes a slice of interface{}, each of which must be []float64
 * arity 2 for x,y coords
 */
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
