package geometry

import (
	"encoding/json"
	"math"
	"os"
)

type Scene struct {
	Width, Height    float64 `json:"-"`
	Rows, Cols       int     `json:"-"`
	Objects          []*Shape
	Camera           Vec3
	Pitch, Yaw, Roll float64
	Near             float64 `json:"-"`
	PixW, PixH       float64 `json:"-"`
}

func ParseScene(filename string, width, height, fov float64, cols, rows int) Scene {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var scene Scene

	err = json.NewDecoder(f).Decode(&scene)
	if err != nil {
		panic(err)
	}

	scene.Near = math.Abs(fov / math.Tan(fov/2.0))
	scene.Width, scene.Height = width, height
	scene.Cols, scene.Rows = cols, rows
	scene.PixW = 2 * width / float64(cols)
	scene.PixH = 2 * height / float64(rows)

	return scene
}
