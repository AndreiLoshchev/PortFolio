package usageoreverse


import (
	"bufio"
    "bytes"
	"math"
	"strconv"
	"strings"
)

const (
	POLYGON        = "POLYGON"
	MULTI_POLYGON  = "MULTIPOLYGON"
	POLYGONS_SPLIT = ")),(("
)

type Point struct {
	x float64
	y float64
}

type CountryPolygon struct {
	CountryCode string
	PointList   []*Point
}

type CountryReverser struct {
	allCountryPolygonInfo []CountryPolygon
}

func trimWithSpecialChar(src string) string {
	src = strings.TrimSpace(src)
	src = strings.Trim(src, "(")
	src = strings.Trim(src, ")")
	return src
}

func stringToListOfPoints(src string) []*Point {
	points := strings.Split(src, ",")
	output := make([]*Point, 0)
	for _, p := range points {
		p = trimWithSpecialChar(p)
		info := strings.Split(p, " ")
		pX, _ := strconv.ParseFloat(info[0], 64)
		pY, _ := strconv.ParseFloat(info[1], 64)

		point := &Point{pX, pY}
		output = append(output, point)
	}

	return output
}

func NewCountryReverser(dataPoints []byte) (*CountryReverser, error) {
	c := new(CountryReverser)
	err := c.load(dataPoints)
	return c, err
}

func (c *CountryReverser) GetCountryCode(longitude, latitude float64) string {
	for _, poly := range c.allCountryPolygonInfo {
		if c.pointInPolygon(longitude, latitude, poly.PointList) {
			return poly.CountryCode
		}
	}
	return ""
}


func (c *CountryReverser) load(data_points []byte) error {
	rd := bufio.NewReader(bytes.NewReader(data_points))
	var line string
	for {
        linestring, err := rd.ReadString('\n')
		line = strings.TrimSpace(linestring)
		info := strings.SplitN(line, "=", 2)
		countryCode := info[0]
		data := info[1]
		pointInfo := strings.SplitN(data, " ", 2)
		polygonType := pointInfo[0]
		data = trimWithSpecialChar(pointInfo[1])
		var polyInfo CountryPolygon
		if POLYGON == strings.ToUpper(polygonType) {
			polyInfo.CountryCode = countryCode
			polyInfo.PointList = stringToListOfPoints(data)
			c.allCountryPolygonInfo = append(c.allCountryPolygonInfo, polyInfo)
		} else if MULTI_POLYGON == strings.ToUpper(polygonType) {
			polygons := strings.Split(data, POLYGONS_SPLIT)
			for _, p := range polygons {
				polyInfo.CountryCode = countryCode
				polyInfo.PointList = stringToListOfPoints(p)
				c.allCountryPolygonInfo = append(c.allCountryPolygonInfo, polyInfo)
			}
		}
        if err != nil {
            break
        }        
	}
	return nil
}


func (c *CountryReverser) pointInPolygon(targetX, targetY float64, pointsList []*Point) bool {
	polyCorners := len(pointsList)
	for _, p := range pointsList {
		if p.x != 0 && int(math.Abs(p.y-targetY)) > 90 {
			return false
		}
		if p.y != 0 && int(math.Abs(p.x-targetX)) > 90 {
			return false
		}
	}
	j := polyCorners - 1
	oddNodes := false
	for i := 0; i < polyCorners; i++ {
		if pointsList[i].y < targetY && pointsList[j].y >= targetY || pointsList[j].y < targetY && pointsList[i].y >= targetY {
			if pointsList[i].x <= targetX || pointsList[j].x <= targetX {
				oddNodes = (oddNodes != (pointsList[i].x+(targetY-pointsList[i].y)/(pointsList[j].y-pointsList[i].y)*(pointsList[j].x-pointsList[i].x) < targetX))
			}
		}
		j = i
	}
	return oddNodes
}