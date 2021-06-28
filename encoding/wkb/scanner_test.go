package wkb

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/spatial-go/geoos"
)

var SRID = []byte{215, 15, 0, 0}

func TestScanNil(t *testing.T) {
	s := Scanner(nil)
	err := s.Scan(testPointData)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}

	if !geoos.Equal(s.Geometry, testPoint) {
		t.Errorf("incorrect geometry: %v != %v", s.Geometry, testPoint)
	}

	t.Run("scan nil data", func(t *testing.T) {
		var p geoos.Point
		s := Scanner(&p)

		err := s.Scan(nil)
		if err != nil {
			t.Errorf("should noop for nil data: %v", err)
		}

		if s.Valid {
			t.Errorf("valid should be false for nil values")
		}
	})

	t.Run("scan nil byte interface", func(t *testing.T) {
		var p geoos.Point
		s := Scanner(&p)

		var b []byte
		err := s.Scan(b)
		if err != nil {
			t.Errorf("should noop for nil data: %v", err)
		}

		if s.Valid {
			t.Errorf("valid should be false for nil values")
		}
	})
}

func TestScanHexData(t *testing.T) {
	cases := []struct {
		name     string
		data     []byte
		expected geoos.Point
	}{
		{
			name:     "point",
			data:     []byte(`\x0101000000e0d57267266e4840b22ac24d46b50240`),
			expected: geoos.Point{48.860547, 2.338513},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var p geoos.Point
			s := Scanner(&p)

			err := s.Scan(tc.data)
			if err != nil {
				t.Fatalf("scan error: %v", err)
			}

			if !p.Equal(tc.expected) {
				t.Errorf("unequal data")
				t.Log(p)
				t.Log(tc.expected)
			}
		})
	}
}

func TestScanPoint(t *testing.T) {
	cases := []struct {
		name     string
		data     []byte
		expected geoos.Point
	}{
		{
			name:     "point",
			data:     testPointData,
			expected: testPoint,
		},
		{
			name:     "point with MySQL SRID",
			data:     append(SRID, testPointData...),
			expected: testPoint,
		},
		{
			name:     "single multi-point",
			data:     testMultiPointSingleData,
			expected: testMultiPointSingle[0],
		},
		{
			name:     "single multi-point with MySQL SRID",
			data:     append(SRID, testMultiPointSingleData...),
			expected: testMultiPointSingle[0],
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var p geoos.Point
			s := Scanner(&p)
			err := s.Scan(tc.data)
			if err != nil {
				t.Fatalf("scan error: %v", err)
			}

			if !p.Equal(tc.expected) {
				t.Errorf("unequal data")
				t.Log(p)
				t.Log(tc.expected)
			}

			if p.Equal(s.Geometry) {
				t.Errorf("should set to scanner's geometry attribute")
			}

			if !s.Valid {
				t.Errorf("should set valid to true")
			}
		})
	}
}

func TestScanPoint_Errors(t *testing.T) {
	// error conditions
	cases := []struct {
		name string
		data interface{}
		err  error
	}{
		{
			name: "incorrect data",
			data: 123,
			err:  ErrUnsupportedDataType,
		},
		{
			name: "not wkb",
			data: []byte{0, 0, 0, 0, 1, 192, 94, 157, 24, 227, 60, 152, 15, 64, 66, 222, 128, 39},
			err:  ErrNotWKB,
		},
		{
			name: "invalid first byte",
			data: []byte{3, 1, 0, 0, 0, 15, 152, 60, 227, 24, 157, 94, 192, 205, 11, 17, 39, 128, 222, 66, 64},
			err:  ErrNotWKB,
		},
		{
			name: "incorrect geometry",
			data: testLineStringData,
			err:  ErrIncorrectGeometry,
		},
		{
			name: "incorrect geometry with MySQL SRID",
			data: append(SRID, testLineStringData...),
			err:  ErrIncorrectGeometry,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var p geoos.Point
			s := Scanner(&p)
			err := s.Scan(tc.data)
			if err != tc.err {
				t.Errorf("incorrect error: %v != %v", err, tc.err)
			}

			if s.Geometry != nil {
				t.Errorf("geometry should be nil on errors")
			}

			if s.Valid {
				t.Errorf("valid should be false on errors")
			}
		})
	}
}

func TestScanMultiPoint(t *testing.T) {
	cases := []struct {
		name     string
		data     []byte
		expected geoos.MultiPoint
	}{
		{
			name:     "multi point",
			data:     testMultiPointData,
			expected: testMultiPoint,
		},
		{
			name:     "multi point with MySQL SRID",
			data:     append(SRID, testMultiPointData...),
			expected: testMultiPoint,
		},
		{
			name:     "point should covert to multi point",
			data:     testPointData,
			expected: geoos.MultiPoint{testPoint},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var mp geoos.MultiPoint
			s := Scanner(&mp)
			err := s.Scan(tc.data)
			if err != nil {
				t.Fatalf("scan error: %v", err)
			}

			if !mp.Equal(tc.expected) {
				t.Errorf("unequal data")
				t.Log(mp)
				t.Log(tc.expected)
			}

			if !reflect.DeepEqual(mp, s.Geometry) {
				t.Errorf("should set to scanner's geometry attribute")
			}

			if !s.Valid {
				t.Errorf("should set valid to true")
			}
		})
	}
}

func TestScanMultiPoint_Errors(t *testing.T) {
	cases := []struct {
		name string
		data interface{}
		err  error
	}{
		{
			name: "does not like line string",
			data: testLineStringData,
			err:  ErrIncorrectGeometry,
		},
		{
			name: "incorrect data",
			data: 123,
			err:  ErrUnsupportedDataType,
		},
		{
			name: "not wkb",
			data: []byte{0, 0, 0, 0, 1, 192, 94},
			err:  ErrNotWKB,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var mp geoos.MultiPoint
			s := Scanner(&mp)
			err := s.Scan(tc.data)
			if err != tc.err {
				t.Errorf("incorrect error: %v != %v", err, tc.err)
			}

			if s.Geometry != nil {
				t.Errorf("geometry should be nil on errors")
			}

			if s.Valid {
				t.Errorf("valid should be false on errors")
			}
		})
	}
}

func TestScanLineString(t *testing.T) {
	cases := []struct {
		name     string
		data     []byte
		expected geoos.LineString
	}{
		{
			name:     "line string",
			data:     testLineStringData,
			expected: testLineString,
		},
		{
			name:     "line string with MySQL SRID",
			data:     append(SRID, testLineStringData...),
			expected: testLineString,
		},
		{
			name:     "single multi line string",
			data:     testMultiLineStringSingleData,
			expected: testMultiLineStringSingle[0],
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var ls geoos.LineString
			s := Scanner(&ls)
			err := s.Scan(tc.data)
			if err != nil {
				t.Fatalf("scan error: %v", err)
			}

			if !ls.Equal(tc.expected) {
				t.Errorf("unequal data")
				t.Log(ls)
				t.Log(tc.expected)
			}

			if !reflect.DeepEqual(ls, s.Geometry) {
				t.Errorf("should set to scanner's geometry attribute")
			}

			if !s.Valid {
				t.Errorf("should set valid to true")
			}
		})
	}
}

func TestScanLineString_Errors(t *testing.T) {
	cases := []struct {
		name string
		data interface{}
		err  error
	}{
		{
			name: "does not like multi point",
			data: testMultiPointData,
			err:  ErrIncorrectGeometry,
		},
		{
			name: "incorrect data",
			data: 123,
			err:  ErrUnsupportedDataType,
		},
		{
			name: "not wkb",
			data: []byte{0, 0, 0, 0, 2, 192, 94},
			err:  ErrNotWKB,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var ls geoos.LineString
			s := Scanner(&ls)
			err := s.Scan(tc.data)
			if err != tc.err {
				t.Errorf("incorrect error: %v != %v", err, tc.err)
			}

			if s.Geometry != nil {
				t.Errorf("geometry should be nil on errors")
			}

			if s.Valid {
				t.Errorf("valid should be false on errors")
			}
		})
	}
}

func TestScanMultiLineString(t *testing.T) {
	cases := []struct {
		name     string
		data     []byte
		expected geoos.MultiLineString
	}{
		{
			name:     "line string",
			data:     testLineStringData,
			expected: geoos.MultiLineString{testLineString},
		},
		{
			name:     "multi line string",
			data:     testMultiLineStringData,
			expected: testMultiLineString,
		},
		{
			name:     "multi line string with MySQL SRID",
			data:     append(SRID, testMultiLineStringData...),
			expected: testMultiLineString,
		},
		{
			name:     "single multi line string",
			data:     testMultiLineStringSingleData,
			expected: testMultiLineStringSingle,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var mls geoos.MultiLineString
			s := Scanner(&mls)
			err := s.Scan(tc.data)
			if err != nil {
				t.Fatalf("scan error: %v", err)
			}

			if !mls.Equal(tc.expected) {
				t.Errorf("unequal data")
				t.Log(mls)
				t.Log(tc.expected)
			}

			if !reflect.DeepEqual(mls, s.Geometry) {
				t.Errorf("should set to scanner's geometry attribute")
			}

			if !s.Valid {
				t.Errorf("should set valid to true")
			}
		})
	}
}

func TestScanMultiLineString_Errors(t *testing.T) {
	cases := []struct {
		name string
		data interface{}
		err  error
	}{
		{
			name: "does not like multi point",
			data: testMultiPointData,
			err:  ErrIncorrectGeometry,
		},
		{
			name: "incorrect data",
			data: 123,
			err:  ErrUnsupportedDataType,
		},
		{
			name: "not wkb",
			data: []byte{0, 0, 0, 0, 5, 192, 94},
			err:  ErrNotWKB,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var mls geoos.MultiLineString
			s := Scanner(&mls)
			err := s.Scan(tc.data)
			if err != tc.err {
				t.Errorf("incorrect error: %v != %v", err, tc.err)
			}

			if s.Geometry != nil {
				t.Errorf("geometry should be nil on errors")
			}

			if s.Valid {
				t.Errorf("valid should be false on errors")
			}
		})
	}
}

func TestScanRing(t *testing.T) {
	cases := []struct {
		name     string
		data     []byte
		expected geoos.Ring
	}{
		{
			name:     "polygon",
			data:     testPolygonData,
			expected: testPolygon[0],
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var r geoos.Ring
			s := Scanner(&r)
			err := s.Scan(tc.data)
			if err != nil {
				t.Fatalf("scan error: %v", err)
			}

			if !r.Equal(tc.expected) {
				t.Errorf("unequal data")
				t.Log(r)
				t.Log(tc.expected)
			}

			if !reflect.DeepEqual(r, s.Geometry) {
				t.Errorf("should set to scanner's geometry attribute")
			}

			if !s.Valid {
				t.Errorf("should set valid to true")
			}
		})
	}
}

func TestScanRing_Errors(t *testing.T) {
	cases := []struct {
		name string
		data interface{}
		err  error
	}{
		{
			name: "does not like line strings",
			data: testLineStringData,
			err:  ErrIncorrectGeometry,
		},
		{
			name: "incorrect data",
			data: 123,
			err:  ErrUnsupportedDataType,
		},
		{
			name: "not wkb",
			data: []byte{0, 0, 0, 0, 1, 192, 94},
			err:  ErrNotWKB,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var r geoos.Ring
			s := Scanner(&r)
			err := s.Scan(tc.data)
			if err != tc.err {
				t.Errorf("incorrect error: %v != %v", err, tc.err)
			}

			if s.Geometry != nil {
				t.Errorf("geometry should be nil on errors")
			}

			if s.Valid {
				t.Errorf("valid should be false on errors")
			}
		})
	}
}

func TestScanPolygon(t *testing.T) {
	cases := []struct {
		name     string
		data     []byte
		expected geoos.Polygon
	}{
		{
			name:     "polygon",
			data:     testPolygonData,
			expected: testPolygon,
		},
		{
			name:     "polygon with MySQL SRID",
			data:     append(SRID, testPolygonData...),
			expected: testPolygon,
		},
		{
			name:     "single multi polygon",
			data:     testMultiPolygonSingleData,
			expected: testMultiPolygonSingle[0],
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var p geoos.Polygon
			s := Scanner(&p)
			err := s.Scan(tc.data)
			if err != nil {
				t.Fatalf("scan error: %v", err)
			}

			if !p.Equal(tc.expected) {
				t.Errorf("unequal data")
				t.Log(p)
				t.Log(tc.expected)
			}

			if !reflect.DeepEqual(p, s.Geometry) {
				t.Errorf("should set to scanner's geometry attribute")
			}

			if !s.Valid {
				t.Errorf("should set valid to true")
			}
		})
	}
}

func TestScanPolygon_Errors(t *testing.T) {
	cases := []struct {
		name string
		data interface{}
		err  error
	}{
		{
			name: "does not like line strings",
			data: testLineStringData,
			err:  ErrIncorrectGeometry,
		},
		{
			name: "incorrect data",
			data: 123,
			err:  ErrUnsupportedDataType,
		},
		{
			name: "not wkb",
			data: []byte{0, 0, 0, 0, 3, 192, 94},
			err:  ErrNotWKB,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var p geoos.Polygon
			s := Scanner(&p)
			err := s.Scan(tc.data)
			if err != tc.err {
				t.Errorf("incorrect error: %v != %v", err, tc.err)
			}

			if s.Geometry != nil {
				t.Errorf("geometry should be nil on errors")
			}

			if s.Valid {
				t.Errorf("valid should be false on errors")
			}
		})
	}
}

func TestScanMultiPolygon(t *testing.T) {
	cases := []struct {
		name     string
		data     []byte
		expected geoos.MultiPolygon
	}{
		{
			name:     "multi polygon",
			data:     testMultiPolygonData,
			expected: testMultiPolygon,
		},
		{
			name:     "multi polygon with MySQL SRID",
			data:     append(SRID, testMultiPolygonData...),
			expected: testMultiPolygon,
		},
		{
			name:     "single multi polygon",
			data:     testMultiPolygonSingleData,
			expected: testMultiPolygonSingle,
		},
		{
			name:     "polygon",
			data:     testPolygonData,
			expected: geoos.MultiPolygon{testPolygon},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var mp geoos.MultiPolygon
			s := Scanner(&mp)
			err := s.Scan(tc.data)
			if err != nil {
				t.Fatalf("scan error: %v", err)
			}

			if !mp.Equal(tc.expected) {
				t.Errorf("unequal data")
				t.Log(mp)
				t.Log(tc.expected)
			}

			if !reflect.DeepEqual(mp, s.Geometry) {
				t.Errorf("should set to scanner's geometry attribute")
			}

			if !s.Valid {
				t.Errorf("should set valid to true")
			}
		})
	}
}

func TestScanMultiPolygon_Errors(t *testing.T) {
	cases := []struct {
		name string
		data interface{}
		err  error
	}{
		{
			name: "does not like line strings",
			data: testLineStringData,
			err:  ErrIncorrectGeometry,
		},
		{
			name: "incorrect data",
			data: 123,
			err:  ErrUnsupportedDataType,
		},
		{
			name: "not wkb",
			data: []byte{0, 0, 0, 0, 6, 192, 94},
			err:  ErrNotWKB,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var mp geoos.MultiPolygon
			s := Scanner(&mp)
			err := s.Scan(tc.data)
			if err != tc.err {
				t.Errorf("incorrect error: %v != %v", err, tc.err)
			}

			if s.Geometry != nil {
				t.Errorf("geometry should be nil on errors")
			}

			if s.Valid {
				t.Errorf("valid should be false on errors")
			}
		})
	}
}

func TestScanCollection(t *testing.T) {
	cases := []struct {
		name     string
		data     []byte
		expected geoos.Collection
	}{
		{
			name:     "collection",
			data:     testCollectionData,
			expected: testCollection,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var c geoos.Collection
			s := Scanner(&c)
			err := s.Scan(tc.data)
			if err != nil {
				t.Fatalf("scan error: %v", err)
			}

			if !c.Equal(tc.expected) {
				t.Errorf("unequal data")
				t.Log(c)
				t.Log(tc.expected)
			}

			if !reflect.DeepEqual(c, s.Geometry) {
				t.Errorf("should set to scanner's geometry attribute")
			}

			if !s.Valid {
				t.Errorf("should set valid to true")
			}
		})
	}
}

func TestScanCollection_Errors(t *testing.T) {
	cases := []struct {
		name string
		data interface{}
		err  error
	}{
		{
			name: "does not like line strings",
			data: testLineStringData,
			err:  ErrIncorrectGeometry,
		},
		{
			name: "incorrect data",
			data: 123,
			err:  ErrUnsupportedDataType,
		},
		{
			name: "not wkb",
			data: []byte{0, 0, 0, 0, 7, 192, 94},
			err:  ErrNotWKB,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var c geoos.Collection
			s := Scanner(&c)
			err := s.Scan(tc.data)
			if err != tc.err {
				t.Errorf("incorrect error: %v != %v", err, tc.err)
			}

			if s.Geometry != nil {
				t.Errorf("geometry should be nil on errors")
			}

			if s.Valid {
				t.Errorf("valid should be false on errors")
			}
		})
	}
}

func TestScanBound(t *testing.T) {
	cases := []struct {
		name     string
		data     []byte
		expected geoos.Bound
	}{
		{
			name:     "point",
			data:     testPointData,
			expected: testPoint.Bound(),
		},
		{
			name:     "multi point",
			data:     testMultiPointData,
			expected: testMultiPoint.Bound(),
		},
		{
			name:     "single multi point",
			data:     testMultiPointSingleData,
			expected: testMultiPointSingle.Bound(),
		},
		{
			name:     "linestring",
			data:     testLineStringData,
			expected: testLineString.Bound(),
		},
		{
			name:     "multi linestring",
			data:     testMultiLineStringData,
			expected: testMultiLineString.Bound(),
		},
		{
			name:     "single multi linestring",
			data:     testMultiLineStringSingleData,
			expected: testMultiLineStringSingle.Bound(),
		},
		{
			name:     "polygon",
			data:     testPolygonData,
			expected: testPolygon.Bound(),
		},
		{
			name:     "multi polygon",
			data:     testMultiPolygonData,
			expected: testMultiPolygon.Bound(),
		},
		{
			name:     "single multi polygon",
			data:     testMultiPolygonSingleData,
			expected: testMultiPolygonSingle.Bound(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var b geoos.Bound
			s := Scanner(&b)
			err := s.Scan(tc.data)
			if err != nil {
				t.Fatalf("scan error: %v", err)
			}

			if !b.Equal(tc.expected) {
				t.Errorf("unequal data")
				t.Log(b)
				t.Log(tc.expected)
			}

			if !reflect.DeepEqual(b, s.Geometry) {
				t.Errorf("should set to scanner's geometry attribute")
			}

			if !s.Valid {
				t.Errorf("should set valid to true")
			}
		})
	}
}

func TestScanBound_Errors(t *testing.T) {
	cases := []struct {
		name string
		data interface{}
		err  error
	}{
		{
			name: "incorrect data",
			data: 123,
			err:  ErrUnsupportedDataType,
		},
		{
			name: "not wkb",
			data: []byte{0, 0, 0, 0, 1, 192, 94},
			err:  ErrNotWKB,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var b geoos.Bound
			s := Scanner(&b)
			err := s.Scan(tc.data)
			if err != tc.err {
				t.Errorf("incorrect error: %v != %v", err, tc.err)
			}

			if s.Geometry != nil {
				t.Errorf("geometry should be nil on errors")
			}

			if s.Valid {
				t.Errorf("valid should be false on errors")
			}
		})
	}
}

func TestValue(t *testing.T) {
	t.Run("marshalls geometry", func(t *testing.T) {
		val, err := Value(testPoint).Value()
		if err != nil {
			t.Errorf("value error: %v", err)
		}

		if !bytes.Equal(val.([]byte), testPointData) {
			t.Errorf("incorrect marshal")
			t.Log(val)
			t.Log(testPointData)
		}
	})

	t.Run("nil value in should set nil value", func(t *testing.T) {
		val, err := Value(nil).Value()
		if err != nil {
			t.Errorf("value error: %v", err)
		}

		if val != nil {
			t.Errorf("should be nil value: %[1]T, %[1]v", val)
		}
	})
}

func TestValue_nil(t *testing.T) {
	var (
		mp    geoos.MultiPoint
		ls    geoos.LineString
		mls   geoos.MultiLineString
		r     geoos.Ring
		poly  geoos.Polygon
		mpoly geoos.MultiPolygon
		c     geoos.Collection
	)

	cases := []struct {
		name string
		geom geoos.Geometry
	}{
		{
			name: "nil multi point",
			geom: mp,
		},
		{
			name: "nil line string",
			geom: ls,
		},
		{
			name: "nil multi line string",
			geom: mls,
		},
		{
			name: "nil ring",
			geom: r,
		},
		{
			name: "nil polygon",
			geom: poly,
		},
		{
			name: "nil multi polygon",
			geom: mpoly,
		},
		{
			name: "nil collection",
			geom: c,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := Value(tc.geom).Value()
			if err != nil {
				t.Errorf("value error: %v", err)
			}

			if val != nil {
				t.Errorf("should be nil value: %[1]T, %[1]v", val)
			}
		})
	}
}

func BenchmarkScan_point(b *testing.B) {
	p := geoos.Point{1, 2}
	data, err := Marshal(p)
	if err != nil {
		b.Fatal(err)
	}

	var r geoos.Point
	s := Scanner(&r)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := s.Scan(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecode_point(b *testing.B) {
	p := geoos.Point{1, 2}
	data, err := Marshal(p)
	if err != nil {
		b.Fatal(err)
	}

	r := bytes.NewReader(data)
	d := NewDecoder(r)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := d.Decode()
		if err != nil {
			b.Fatal(err)
		}

		r.Reset(data)
	}
}

func BenchmarkScan_lineString(b *testing.B) {
	var ls geoos.LineString
	for i := 0; i < 100; i++ {
		ls = append(ls, geoos.Point{float64(i), float64(i)})
	}
	data, err := Marshal(ls)
	if err != nil {
		b.Fatal(err)
	}

	var r geoos.LineString
	s := Scanner(&r)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := s.Scan(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecode_lineString(b *testing.B) {
	var ls geoos.LineString
	for i := 0; i < 100; i++ {
		ls = append(ls, geoos.Point{float64(i), float64(i)})
	}
	data, err := Marshal(ls)
	if err != nil {
		b.Fatal(err)
	}

	r := bytes.NewReader(data)
	d := NewDecoder(r)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := d.Decode()
		if err != nil {
			b.Fatal(err)
		}

		r.Reset(data)
	}
}

func BenchmarkScan_multiLineString(b *testing.B) {
	var mls geoos.MultiLineString
	for i := 0; i < 10; i++ {
		var ls geoos.LineString
		for j := 0; j < 100; j++ {
			ls = append(ls, geoos.Point{float64(i), float64(i)})
		}
		mls = append(mls, ls)
	}
	data, err := Marshal(mls)
	if err != nil {
		b.Fatal(err)
	}

	var r geoos.MultiLineString
	s := Scanner(&r)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := s.Scan(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecode_multiLineString(b *testing.B) {
	var mls geoos.MultiLineString
	for i := 0; i < 10; i++ {
		var ls geoos.LineString
		for j := 0; j < 100; j++ {
			ls = append(ls, geoos.Point{float64(i), float64(i)})
		}
		mls = append(mls, ls)
	}
	data, err := Marshal(mls)
	if err != nil {
		b.Fatal(err)
	}

	r := bytes.NewReader(data)
	d := NewDecoder(r)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := d.Decode()
		if err != nil {
			b.Fatal(err)
		}

		r.Reset(data)
	}
}

func BenchmarkScan_polygon(b *testing.B) {
	var p geoos.Polygon
	for i := 0; i < 1; i++ {
		var r geoos.Ring
		for j := 0; j < 6; j++ {
			r = append(r, geoos.Point{float64(i), float64(i)})
		}
		p = append(p, r)
	}
	data, err := Marshal(p)
	if err != nil {
		b.Fatal(err)
	}

	var r geoos.Polygon
	s := Scanner(&r)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := s.Scan(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecode_polygon(b *testing.B) {
	var p geoos.Polygon
	for i := 0; i < 1; i++ {
		var r geoos.Ring
		for j := 0; j < 6; j++ {
			r = append(r, geoos.Point{float64(i), float64(i)})
		}
		p = append(p, r)
	}
	data, err := Marshal(p)
	if err != nil {
		b.Fatal(err)
	}

	r := bytes.NewReader(data)
	d := NewDecoder(r)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := d.Decode()
		if err != nil {
			b.Fatal(err)
		}

		r.Reset(data)
	}
}
