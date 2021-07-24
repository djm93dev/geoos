package relate

import (
	"github.com/spatial-go/geoos/algorithm/matrix"
)

// PolygonRelate  be used during the relate computation.
type PolygonRelate struct {
	matrix.PolygonMatrix
	other matrix.Steric
}

// IntersectionMatrix Gets the IntersectionMatrix for the spatial relationship
// between the input geometries.
func (p *PolygonRelate) IntersectionMatrix(im *matrix.IntersectionMatrix) *matrix.IntersectionMatrix {
	switch p.other.(type) {
	case matrix.Matrix:
		pr := &PointRelate{p.other.(matrix.Matrix), p.PolygonMatrix}
		return pr.IntersectionMatrix(im).Transpose()
	case matrix.LineMatrix:
		lr := &LineRelate{p.other.(matrix.LineMatrix), p.PolygonMatrix}
		return lr.IntersectionMatrix(im).Transpose()
	case matrix.PolygonMatrix:
		p.computePolygon(im)
		return im
	}
	return im
}

func (p *PolygonRelate) computePolygon(im *matrix.IntersectionMatrix) {
	inRing := -1
	for i, v := range p.other.(matrix.PolygonMatrix) {
		l := p.PolygonMatrix[0]
		if IsIntersectionEdge(l, v) {
			inRing = 1
			break
		}
		if i == 0 {
			if InPolygon(l[0], v) {
				inRing = 0
			} else {
				inRing = 2
			}
		} else {

			if InPolygon(l[0], matrix.LineMatrix(v)) {
				if inRing != 2 {
					inRing = 2
					break
				}
			}
		}

	}
	switch inRing {
	case 0:
		im.SetAtLeastString("2FF1FF212")
	case 1:
		im.SetAtLeastString("212101212")
	case 2:
		im.SetAtLeastString("FF2FF1212")
	}
}
