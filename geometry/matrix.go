package geometry

import (
	"math"
)

type Mat4 [4][4]float64

func (m Mat4) MultMat4(m2 Mat4) Mat4 {
	return Mat4{
		{
			m[0][0]*m2[0][0] + m[0][1]*m2[1][0] + m[0][2]*m2[2][0] + m[0][3]*m2[3][0],
			m[0][0]*m2[0][1] + m[0][1]*m2[1][1] + m[0][2]*m2[2][1] + m[0][3]*m2[3][1],
			m[0][0]*m2[0][2] + m[0][1]*m2[1][2] + m[0][2]*m2[2][2] + m[0][3]*m2[3][2],
			m[0][0]*m2[0][3] + m[0][1]*m2[1][3] + m[0][2]*m2[2][3] + m[0][3]*m2[3][3],
		}, {
			m[1][0]*m2[0][0] + m[1][1]*m2[1][0] + m[1][2]*m2[2][0] + m[1][3]*m2[3][0],
			m[1][0]*m2[0][1] + m[1][1]*m2[1][1] + m[1][2]*m2[2][1] + m[1][3]*m2[3][1],
			m[1][0]*m2[0][2] + m[1][1]*m2[1][2] + m[1][2]*m2[2][2] + m[1][3]*m2[3][2],
			m[1][0]*m2[0][3] + m[1][1]*m2[1][3] + m[1][2]*m2[2][3] + m[1][3]*m2[3][3],
		}, {
			m[2][0]*m2[0][0] + m[2][1]*m2[1][0] + m[2][2]*m2[2][0] + m[2][3]*m2[3][0],
			m[2][0]*m2[0][1] + m[2][1]*m2[1][1] + m[2][2]*m2[2][1] + m[2][3]*m2[3][1],
			m[2][0]*m2[0][2] + m[2][1]*m2[1][2] + m[2][2]*m2[2][2] + m[2][3]*m2[3][2],
			m[2][0]*m2[0][3] + m[2][1]*m2[1][3] + m[2][2]*m2[2][3] + m[2][3]*m2[3][3],
		}, {
			m[3][0]*m2[0][0] + m[3][1]*m2[1][0] + m[3][2]*m2[2][0] + m[3][3]*m2[3][0],
			m[3][0]*m2[0][1] + m[3][1]*m2[1][1] + m[3][2]*m2[2][1] + m[3][3]*m2[3][1],
			m[3][0]*m2[0][2] + m[3][1]*m2[1][2] + m[3][2]*m2[2][2] + m[3][3]*m2[3][2],
			m[3][0]*m2[0][3] + m[3][1]*m2[1][3] + m[3][2]*m2[2][3] + m[3][3]*m2[3][3],
		},
	}
}

func (m Mat4) MultVec3(v Vec3) Vec3 {
	return Vec3{
		m[0][0]*v.X + m[0][1]*v.Y + m[0][2]*v.Z,
		m[1][0]*v.X + m[1][1]*v.Y + m[1][2]*v.Z,
		m[2][0]*v.X + m[2][1]*v.Y + m[2][2]*v.Z,
	}
}

func RotateVector(angle float64, axis, vec Vec3) Vec3 {
	sin, cos := math.Sin(angle), math.Cos(angle)
	x, y, z := axis.X, axis.Y, axis.Z
	m := Mat4{
		{cos + x*x*(1-cos), x*y*(1-cos) - x*sin, x*z*(1-cos) + y*sin, 0},
		{y*x*(1-cos) + z*sin, cos + y*y*(1-cos), y*z*(1-cos) - x*sin, 0},
		{z*x*(1-cos) - y*sin, z*y*(1-cos) + x*sin, cos + z*z*(1-cos), 0},
		{0, 0, 0, 1},
	}

	return m.MultVec3(vec)
}

func PitchYawRollVector(pitch, yaw, roll float64, vec Vec3) Vec3 {
	s, c := math.Sin(pitch), math.Cos(pitch)
	x := Mat4{
		{1, 0, 0, 0},
		{0, c, s, 0},
		{0, -s, c, 0},
		{0, 0, 0, 1},
	}

	s, c = math.Sin(yaw), math.Cos(yaw)
	y := Mat4{
		{c, 0, -s, 0},
		{0, 1, 0, 0},
		{s, 0, c, 0},
		{0, 0, 0, 1},
	}

	s, c = math.Sin(roll), math.Cos(roll)
	z := Mat4{
		{c, s, 0, 0},
		{-s, c, 0, 0},
		{0, 0, 1, 0},
		{0, 0, 0, 1},
	}

	return z.MultMat4(x).MultMat4(y).MultVec3(vec)
}
