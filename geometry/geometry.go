package geometry

import (
	"errors"
	"math"
)

var ErrShapeType = errors.New("invalid shape type")
var ErrMaterial = errors.New("invalid material")

type Shape struct {
	Type     ShapeType
	Material Material
	Color    Vec3
	Emission Vec3
	Position Vec3
	Normal   Vec3
	Radius   float64
}

type ShapeType int

const (
	kindSphere ShapeType = iota
	kindPlane
	kindCube
)

var shapeTypes = map[string]ShapeType{
	"\"SPHERE\"": kindSphere,
	"\"PLANE\"":  kindPlane,
	"\"CUBE\"":   kindCube,
}
var shapeTypesReverse = map[ShapeType]string{
	kindSphere: "\"SPHERE\"",
	kindPlane:  "\"PLANE\"",
	kindCube:   "\"CUBE\"",
}

func (t *ShapeType) MarshalJSON() ([]byte, error) {
	return []byte(shapeTypesReverse[*t]), nil
}

func (t *ShapeType) UnmarshalJSON(b []byte) error {
	v, ok := shapeTypes[string(b)]
	if !ok {
		return ErrShapeType
	}
	*t = v
	return nil
}

type Material int

const (
	DIFFUSE Material = iota
	SPECULAR
	REFRACTIVE
)

var materials = map[string]Material{
	"\"DIFFUSE\"":    DIFFUSE,
	"\"SPECULAR\"":   SPECULAR,
	"\"REFRACTIVE\"": REFRACTIVE,
}
var materialsReverse = map[Material]string{
	DIFFUSE:    "\"DIFFUSE\"",
	SPECULAR:   "\"SPECULAR\"",
	REFRACTIVE: "\"REFRACTIVE\"",
}

func (t *Material) MarshalJSON() ([]byte, error) {
	return []byte(materialsReverse[*t]), nil
}

func (t *Material) UnmarshalJSON(b []byte) error {
	v, ok := materials[string(b)]
	if !ok {
		return ErrMaterial
	}
	*t = v
	return nil
}

func (s *Shape) Intersects(ray *Ray) float64 {
	switch s.Type {
	case kindSphere:
		return sphereIntersects(s, ray)
	case kindPlane:
		return planeIntersects(s, ray)
	case kindCube:
		return cubeIntersects(s, ray)
	}
	panic("unreachable")
}

func (s *Shape) NormalDir(point Vec3) Vec3 {
	switch s.Type {
	case kindSphere:
		return sphereNormal(s, point)
	case kindPlane:
		return planeNormal(s, point)
	case kindCube:
		return cubeNormal(s, point)
	}
	panic("unreachable")
}

var positiveInfinity = math.Inf(+1)

func Plane(position, emission, color, normal Vec3, materialType Material) *Shape {
	return &Shape{
		Type:     kindPlane,
		Material: materialType,
		Color:    color,
		Emission: emission,
		Position: position,
		Normal:   normal,
	}
}

func Sphere(radius float64, position, emission, color Vec3, materialType Material) *Shape {
	return &Shape{
		Type:     kindSphere,
		Material: materialType,
		Color:    color,
		Emission: emission,
		Position: position,
		Radius:   radius,
	}
}

func Cube(radius float64, position, emission, color Vec3, materialType Material) *Shape {
	return &Shape{
		Type:     kindCube,
		Material: materialType,
		Color:    color,
		Emission: emission,
		Position: position,
		Radius:   radius,
	}
}

func intersectPlane(origin, normal Vec3, r *Ray) float64 {
	// Orthogonal
	dot := r.Direction.Dot(normal)
	if dot == 0 {
		return positiveInfinity
	}
	return origin.SubDot(r.Origin, normal) / dot
}

func planeIntersects(s *Shape, r *Ray) float64 {
	return intersectPlane(s.Position, s.Normal, r)
}

func sphereIntersects(s *Shape, ray *Ray) float64 {
	difference := s.Position.Sub(ray.Origin)
	dot := difference.Dot(ray.Direction)
	hypotenuse := dot*dot - difference.Dot(difference) + s.Radius*s.Radius

	if hypotenuse < 0 {
		return positiveInfinity
	}

	hypotenuse = math.Sqrt(hypotenuse)
	if diff := dot - hypotenuse; diff > 0 {
		return diff
	}
	if diff := dot + hypotenuse; diff > 0 {
		return diff
	}
	return positiveInfinity
}

func cubeIntersects(s *Shape, r *Ray) float64 {
	// TODO: optimize this heavily
	min := positiveInfinity
	for i := 0; i < 6; i++ {
		var normal Vec3
		switch i {
		case 0:
			normal.X = -s.Radius
		case 1:
			normal.X = s.Radius
		case 2:
			normal.Y = -s.Radius
		case 3:
			normal.Y = s.Radius
		case 4:
			normal.Z = -s.Radius
		case 5:
			normal.Z = s.Radius
		}
		dist := intersectPlane(s.Position.Add(normal), normal, r)
		if dist > 0 && dist < min {
			diff := r.Origin.Add(r.Direction.Mult(dist)).Sub(s.Position)
			if -s.Radius <= diff.X && diff.X <= s.Radius &&
				-s.Radius <= diff.Y && diff.Y <= s.Radius &&
				-s.Radius <= diff.Z && diff.Z <= s.Radius {
				min = dist
			}
		}
	}

	return min
}

func planeNormal(s *Shape, point Vec3) Vec3 {
	return s.Normal
}

func sphereNormal(s *Shape, point Vec3) Vec3 {
	return point.Sub(s.Position)
}

func cubeNormal(s *Shape, point Vec3) Vec3 {
	// TODO: optimize this heavily
	var max float64
	var bestNormal Vec3
	diff := point.Sub(s.Position)
	for i := 0; i < 6; i++ {
		var normal Vec3
		switch i {
		case 0:
			normal.X = -s.Radius
		case 1:
			normal.X = s.Radius
		case 2:
			normal.Y = -s.Radius
		case 3:
			normal.Y = s.Radius
		case 4:
			normal.Z = -s.Radius
		case 5:
			normal.Z = s.Radius
		}
		dot := normal.Dot(diff)
		if dot > max {
			max = dot
			bestNormal = normal
		}
	}

	return bestNormal
}

type Ray struct {
	Origin, Direction Vec3
}
