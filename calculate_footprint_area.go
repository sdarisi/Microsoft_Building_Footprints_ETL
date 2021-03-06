package main

import "io/ioutil"
import "github.com/paulmach/go.geojson"
import "github.com/golang/geo/s2"
import "fmt"
import "sort"
import "os"

//import "math"

func check(e error) {
	if e != nil {
		print("got error")
		panic(e)
	}
}

var geoms map[int]geojson.Geometry
var areas map[int]float64

func parsePolygon(feature geojson.Feature) *s2.Polygon {

	coords := feature.Geometry.Polygon
	var points []s2.Point

	for _, p := range coords[0] {
		points = append(points, s2.PointFromLatLng(s2.LatLngFromDegrees(p[1], p[0])))
	}

	loop := s2.LoopFromPoints(points)
	var loops []*s2.Loop
	loops = append(loops, loop)
	polygon := s2.PolygonFromLoops(loops)
	return polygon
}

func calcArea(polygon s2.Polygon) float64 {
	var area float64 = polygon.Area() * 85011012.19 * 1000 * 1000
	return area
}

type KeyAreaPair struct {
	Key  int
	Area float64
}

type KeyAreaPairList []KeyAreaPair

func (p KeyAreaPairList) Len() int           { return len(p) }
func (p KeyAreaPairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p KeyAreaPairList) Less(i, j int) bool { return p[i].Area < p[j].Area }

func sortAreas(areas map[int]float64) []int {
	list := make(KeyAreaPairList, len(areas))
	i := 0
	for k, v := range areas {
		list[i] = KeyAreaPair{k, v}
		i++
	}
	var keys []int
	sort.Sort(list)
	for _, pair := range list {
		keys = append(keys, pair.Key)
	}
	return keys
}

func main() {
	targetFile := os.Args[1]
	outFile := os.Args[2]

	geo, err := ioutil.ReadFile(targetFile)
	check(err)
	resultFC := geojson.NewFeatureCollection()

	println("Reading GeoJSON File")

	fc1, err := geojson.UnmarshalFeatureCollection(geo)
	var indexOffset int = 0
	geoms = make(map[int]geojson.Geometry)
	areas = make(map[int]float64)
	println("parsing geometries")

	noFeatures := len(fc1.Features)

	for index, feature := range fc1.Features {
		if index%10000 == 0 {
			fmt.Printf("Done %d of %d %d%%  \n", index, noFeatures, index*100.0/noFeatures)
		}
		geoms[index+indexOffset] = *feature.Geometry
		//id := feature.ID
		poly := parsePolygon(*feature)
		area := calcArea(*poly)
		feature.SetProperty("area", area)
		resultFC.AddFeature(feature)

		areas[index+indexOffset] = area
	}
	keys := sortAreas(areas)
	println(len(geoms))
	println(keys[len(keys)-1], areas[len(keys)-1])
	rawJSON, err := resultFC.MarshalJSON()
	check(err)

	ioutil.WriteFile(outFile, rawJSON, 0644)

	check(err)

}
