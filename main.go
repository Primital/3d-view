package main

import (
	"image/color"
	"math"
	"sort"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

const FrameLength = 25 * time.Millisecond

type Point struct {
	X, Y, Z float64
}

type Polygon struct {
	Points []Point
	Color  color.Color
}

type Pyramid struct {
	Sides []Polygon
}

type Vector struct {
	X, Y, Z float64
}

type Matrix struct {
	X Vector
	Y Vector
	Z Vector
}

func (p *Polygon) build(imd *imdraw.IMDraw, win *pixelgl.Window) {
	for _, point := range p.Points {
		imd.Color = p.Color
		imd.Push(pixel.V(point.X, point.Y))
	}
	imd.Color = p.Color
	imd.Push(pixel.V(p.Points[0].X, p.Points[0].Y))
	imd.Polygon(1)
	imd.Draw(win)
}

func (p *Polygon) Transform(m Matrix) *Polygon {
	newPoints := make([]Point, len(p.Points))
	for i, point := range p.Points {
		newPoints[i] = Point{
			X: point.X*m.X.X + point.Y*m.Y.X + point.Z*m.Z.X,
			Y: point.X*m.X.Y + point.Y*m.Y.Y + point.Z*m.Z.Y,
			Z: point.X*m.X.Z + point.Y*m.Y.Z + point.Z*m.Z.Z,
		}
	}
	return &Polygon{
		Points: newPoints,
		Color:  p.Color,
	}
}

func computeCentroid(points []Point) Point {
	var sumX, sumY, sumZ float64
	n := float64(len(points))

	for _, p := range points {
		sumX += p.X
		sumY += p.Y
		sumZ += p.Z
	}

	return Point{
		X: sumX / n,
		Y: sumY / n,
		Z: sumZ / n,
	}
}

func NewPolygon(points []Point, color color.Color) *Polygon {
	if len(points) < 3 {
		panic("A polygon must have at least 3 points")
	}
	return &Polygon{
		Points: points,
		Color:  color,
	}
}

func euclideanDistance(a, b Point) float64 {
	return math.Sqrt(math.Pow(b.X-a.X, 2) + math.Pow(b.Y-a.Y, 2) + math.Pow(b.Z-a.Z, 2))
}

type PointsSet []Point

type ByCentroidDistance struct {
	ps       []PointsSet
	refPoint Point
}

func (bd ByCentroidDistance) Len() int {
	return len(bd.ps)
}

func (bd ByCentroidDistance) Swap(i, j int) {
	bd.ps[i], bd.ps[j] = bd.ps[j], bd.ps[i]
}

func (bd ByCentroidDistance) Less(i, j int) bool {

	centroidI := computeCentroid(bd.ps[i])
	centroidJ := computeCentroid(bd.ps[j])

	return euclideanDistance(centroidI, bd.refPoint) > euclideanDistance(centroidJ, bd.refPoint)
}

func NewPyramid(p1, p2, p3, p4 Vector) []Polygon {
	return []Polygon{
		*NewPolygon([]Point{{X: p1.X, Y: p1.Y, Z: p1.Z}, {X: p2.X, Y: p2.Y, Z: p2.Z}, {X: p3.X, Y: p3.Y, Z: p3.Z}}, colornames.Red),
		*NewPolygon([]Point{{X: p1.X, Y: p1.Y, Z: p1.Z}, {X: p2.X, Y: p2.Y, Z: p2.Z}, {X: p4.X, Y: p4.Y, Z: p4.Z}}, colornames.Green),
		*NewPolygon([]Point{{X: p1.X, Y: p1.Y, Z: p1.Z}, {X: p3.X, Y: p3.Y, Z: p3.Z}, {X: p4.X, Y: p4.Y, Z: p4.Z}}, colornames.Blue),
		*NewPolygon([]Point{{X: p2.X, Y: p2.Y, Z: p2.Z}, {X: p3.X, Y: p3.Y, Z: p3.Z}, {X: p4.X, Y: p4.Y, Z: p4.Z}}, colornames.Black),
	}
}

type Space struct {
	Matrix  Matrix
	Objects []*Polygon
}

func NewSpace() *Space {
	return &Space{
		Matrix: Matrix{
			X: Vector{X: 1, Y: 0, Z: 0},
			Y: Vector{X: 0, Y: 1, Z: 0},
			Z: Vector{X: 0, Y: 0, Z: 1},
		},
		Objects: nil,
	}
}

func (s *Space) Rotate(rotationMatrix Matrix) {
	transformedMatrix := Matrix{
		X: Vector{
			X: s.Matrix.X.X*rotationMatrix.X.X + s.Matrix.X.Y*rotationMatrix.Y.X + s.Matrix.X.Z*rotationMatrix.Z.X,
			Y: s.Matrix.X.X*rotationMatrix.X.Y + s.Matrix.X.Y*rotationMatrix.Y.Y + s.Matrix.X.Z*rotationMatrix.Z.Y,
			Z: s.Matrix.X.X*rotationMatrix.X.Z + s.Matrix.X.Y*rotationMatrix.Y.Z + s.Matrix.X.Z*rotationMatrix.Z.Z,
		},
		Y: Vector{
			X: s.Matrix.Y.X*rotationMatrix.X.X + s.Matrix.Y.Y*rotationMatrix.Y.X + s.Matrix.Y.Z*rotationMatrix.Z.X,
			Y: s.Matrix.Y.X*rotationMatrix.X.Y + s.Matrix.Y.Y*rotationMatrix.Y.Y + s.Matrix.Y.Z*rotationMatrix.Z.Y,
			Z: s.Matrix.Y.X*rotationMatrix.X.Z + s.Matrix.Y.Y*rotationMatrix.Y.Z + s.Matrix.Y.Z*rotationMatrix.Z.Z,
		},
		Z: Vector{
			X: s.Matrix.Z.X*rotationMatrix.X.X + s.Matrix.Z.Y*rotationMatrix.Y.X + s.Matrix.Z.Z*rotationMatrix.Z.X,
			Y: s.Matrix.Z.X*rotationMatrix.X.Y + s.Matrix.Z.Y*rotationMatrix.Y.Y + s.Matrix.Z.Z*rotationMatrix.Z.Y,
			Z: s.Matrix.Z.X*rotationMatrix.X.Z + s.Matrix.Z.Y*rotationMatrix.Y.Z + s.Matrix.Z.Z*rotationMatrix.Z.Z,
		},
	}
	s.Matrix = transformedMatrix
}

func (s *Space) RotateX(angle float64) {
	rotationMatrix := Matrix{
		X: Vector{X: 1, Y: 0, Z: 0},
		Y: Vector{X: 0, Y: math.Cos(angle), Z: math.Sin(angle)},
		Z: Vector{X: 0, Y: -math.Sin(angle), Z: math.Cos(angle)},
	}
	s.Rotate(rotationMatrix)
}

func (s *Space) RotateY(angle float64) {
	rotationMatrix := Matrix{
		X: Vector{X: math.Cos(angle), Y: 0, Z: -math.Sin(angle)},
		Y: Vector{X: 0, Y: 1, Z: 0},
		Z: Vector{X: math.Sin(angle), Y: 0, Z: math.Cos(angle)},
	}
	s.Rotate(rotationMatrix)
}

func (s *Space) RotateZ(angle float64) {
	rotationMatrix := Matrix{
		X: Vector{X: math.Cos(angle), Y: math.Sin(angle), Z: 0},
		Y: Vector{X: -math.Sin(angle), Y: math.Cos(angle), Z: 0},
		Z: Vector{X: 0, Y: 0, Z: 1},
	}
	s.Rotate(rotationMatrix)

}

func (s *Space) AddObject(p *Polygon) {
	s.Objects = append(s.Objects, p)
}

func matrixMultiply(m Matrix, p Point) Point {
	return Point{
		X: p.X*m.X.X + p.Y*m.Y.X + p.Z*m.Z.X,
		Y: p.X*m.X.Y + p.Y*m.Y.Y + p.Z*m.Z.Y,
		Z: p.X*m.X.Z + p.Y*m.Y.Z + p.Z*m.Z.Z,
	}
}

func (s *Space) Draw(imd *imdraw.IMDraw, win *pixelgl.Window) {
	referencePoint := Point{X: 0, Y: 0, Z: 200}

	// Sort polygons by distance from the reference point
	polygons := make([]PointsSet, len(s.Objects))
	for i, obj := range s.Objects {
		points := make([]Point, 3)
		for _, point := range obj.Points {
			points = append(points, matrixMultiply(s.Matrix, point))
		}
		polygons[i] = points
	}
	sort.Sort(ByCentroidDistance{ps: polygons, refPoint: referencePoint})
	for _, obj := range s.Objects {
		newObj := obj.Transform(s.Matrix)
		newObj.build(imd, win)
	}
}

func run() {
	space := NewSpace()
	cfg := pixelgl.WindowConfig{
		Title:  "3D view",
		Bounds: pixel.R(-150, -150, 150, 150),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	p1 := Point{X: -50, Y: 0, Z: -50}
	p2 := Point{X: 0, Y: 100, Z: 0}
	p3 := Point{X: 50, Y: 0, Z: -50}
	p4 := Point{X: -50, Y: 0, Z: 50}
	p5 := Point{X: 50, Y: 0, Z: 50}

	// Add axis lines
	// X axis
	space.AddObject(NewPolygon([]Point{
		Point{X: -300, Y: 0, Z: 0},
		Point{X: -300, Y: 0, Z: 0},
		Point{X: 300, Y: 0, Z: 0},
	}, colornames.Black))
	// Y axis
	space.AddObject(NewPolygon([]Point{
		Point{X: 0, Y: -300, Z: 0},
		Point{X: 0, Y: -300, Z: 0},
		Point{X: 0, Y: 300, Z: 0},
	}, colornames.Black))
	// Z axis
	space.AddObject(NewPolygon([]Point{
		Point{X: 0, Y: 0, Z: 300},
		Point{X: 0, Y: 0, Z: 300},
		Point{X: 0, Y: 0, Z: -300},
	}, colornames.Black))

	// floor
	space.AddObject(NewPolygon([]Point{
		p1, p4, p5,
	}, colornames.Black))
	// "left"
	space.AddObject(NewPolygon([]Point{
		p1, p2, p4,
	}, colornames.Aliceblue))
	// "right"
	space.AddObject(NewPolygon([]Point{
		p2, p3, p5,
	}, colornames.Aliceblue))
	// "front
	space.AddObject(NewPolygon([]Point{
		p1, p2, p3,
	}, colornames.Aliceblue))
	// "back"
	space.AddObject(NewPolygon([]Point{
		p2, p4, p5,
	}, colornames.Aliceblue))

	// space.RotateZ(-0.033)
	space.RotateX(-0.33)
	for !win.Closed() {
		win.Clear(colornames.Dimgray)

		imd := imdraw.New(nil) // Create a new immediate-mode drawing context
		space.Draw(imd, win)
		// space.RotateX(-0.033)
		space.RotateY(0.033)
		// space.RotateZ(0.001)
		win.Update() // Update the window
		time.Sleep(FrameLength)
	}
}

func main() {
	pixelgl.Run(run)
}
