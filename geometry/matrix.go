package geometry

import (
	"math"
)

type Mat4 [4][4]float64

func (m Mat4) Mult(v *Vec3) Vec3 {
	return Vec3{
		m[0][0]*v.X + m[0][1]*v.Y + m[0][2]*v.Z,
		m[1][0]*v.X + m[1][1]*v.Y + m[1][2]*v.Z,
		m[2][0]*v.X + m[2][1]*v.Y + m[2][2]*v.Z,
	}
}

func RotateVector(angle float64, axis, vec *Vec3) Vec3 {
	sin, cos := math.Sin(angle), math.Cos(angle)
	x, y, z := axis.X, axis.Y, axis.Z
	m := Mat4{
		{cos + x*x*(1-cos), x*y*(1-cos) - x*sin, x*z*(1-cos) + y*sin, 0},
		{y*x*(1-cos) + z*sin, cos + y*y*(1-cos), y*z*(1-cos) - x*sin, 0},
		{z*x*(1-cos) - y*sin, z*y*(1-cos) + x*sin, cos + z*z*(1-cos), 0},
		{0, 0, 0, 1},
	}

	return m.Mult(vec)
}
