package render

import (
	"fmt"
	"github.com/BenLubar/goray/geometry"
	"github.com/BenLubar/goray/kd"
	"image"
	"image/color"
	"math"
	"math/rand"
	"time"
)

func clearLine() {
	fmt.Print("\r\x1b[K")
}

func ClosestIntersection(shapes []*geometry.Shape, ray geometry.Ray) (*geometry.Shape, float64) {
	var closest *geometry.Shape
	bestHit := math.Inf(+1)
	for _, shape := range shapes {
		if hit := shape.Intersects(&ray); hit > 1e-15 && hit < bestHit {
			bestHit = hit
			closest = shape
		}
	}
	return closest, bestHit
}

type Result struct {
	x, y  int
	color geometry.Vec3
}

const (
	AIR   = 1.0
	GLASS = 1.5
)

var (
	PITCH = &geometry.Vec3{1, 0, 0}
	YAW   = &geometry.Vec3{0, 1, 0}
	ROLL  = &geometry.Vec3{0, 0, 1}
)

func MonteCarloPixel(results chan Result, scene *geometry.Scene, diffuseMap, causticsMap *kd.KDNode, start, rows int, rand *rand.Rand) {
	samples := Config.NumRays

	for y := start; y < start+rows; y++ {
		py := scene.Height - scene.Height*2*float64(y)/float64(scene.Rows)
		for x := 0; x < scene.Cols; x++ {
			px := -scene.Width + scene.Width*2*float64(x)/float64(scene.Cols)
			var colorSamples geometry.Vec3
			if x >= Config.Skip.Left && x < scene.Cols-Config.Skip.Right &&
				y >= Config.Skip.Top && y < scene.Rows-Config.Skip.Bottom {
				for sample := 0; sample < samples; sample++ {
					dy, dx := rand.Float64()*scene.PixH, rand.Float64()*scene.PixW
					direction := geometry.Vec3{
						px + dx,
						py + dy,
						scene.Near,
					}.Normalize()
					direction = geometry.RotateVector(scene.Roll, ROLL, &direction)
					direction = geometry.RotateVector(scene.Pitch, PITCH, &direction)
					direction = geometry.RotateVector(scene.Yaw, YAW, &direction)

					contribution := Radiance(geometry.Ray{scene.Camera, direction}, scene, diffuseMap, causticsMap, 0, 1.0, rand)
					colorSamples.AddInPlace(contribution)
				}
			}
			results <- Result{x, y, colorSamples.Mult(1.0 / float64(samples))}
		}
	}
}

func CorrectColor(x float64) float64 {
	return math.Pow(x, 1.0/Config.GammaFactor)*255 + 0.5
}

func CorrectColors(v geometry.Vec3) geometry.Vec3 {
	v.X = CorrectColor(v.X)
	v.Y = CorrectColor(v.Y)
	v.Z = CorrectColor(v.Z)
	return v
}

func mix(a, b geometry.Vec3, factor float64) geometry.Vec3 {
	a.X -= (b.X - a.X) * factor
	a.Y -= (b.Y - a.Y) * factor
	a.Z -= (b.Z - a.Z) * factor
	return a
}

func BloomFilter(img [][]geometry.Vec3, depth int) [][]geometry.Vec3 {
	data := make([][]geometry.Vec3, len(img))
	for i, _ := range data {
		data[i] = make([]geometry.Vec3, len(img[0]))
	}

	const box_width = 2
	factor := 1.0 / math.Pow(2*box_width+1, 2)

	source := img
	for iteration := 0; iteration < depth; iteration++ {
		for y := box_width; y < len(img)-box_width; y++ {
			for x := box_width; x < len(img[0])-box_width; x++ {
				var color geometry.Vec3
				for dy := -box_width; dy <= box_width; dy++ {
					for dx := -box_width; dx <= box_width; dx++ {
						color.AddInPlace(source[y+dy][x+dx])
					}
				}
				data[y][x] = color.Mult(factor)
			}
		}
		fmt.Printf("\rPost Processing %3.0f%%   \r", 100*float64(iteration)/float64(depth))
		source, data = data, source
	}
	return source
}

var Config struct {
	MinDepth    int
	NumRays     int
	Chunks      int
	GammaFactor float64
	BloomFactor int
	Caustics    int

	Skip struct {
		Top, Left, Right, Bottom int
	}
}

func Render(scene geometry.Scene) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, scene.Cols, scene.Rows))
	pixels := make(chan Result, 128)

	workload := scene.Rows / Config.Chunks

	startTime := time.Now()
	globals, caustics := GenerateMaps(scene.Objects)
	fmt.Println(" Done!")
	fmt.Printf("Diffuse Map depth: %v Caustics Map depth: %v\n", globals.Depth(), caustics.Depth())
	fmt.Printf("Photon Maps Done. Generation took: %v\n", time.Since(startTime))

	startTime = time.Now()
	for y := 0; y < scene.Rows; y += workload {
		go MonteCarloPixel(pixels, &scene, globals, caustics, y, workload, rand.New(rand.NewSource(rand.Int63())))
	}

	// Write targets for after effects
	data := make([][]geometry.Vec3, scene.Rows)
	peaks := make([][]geometry.Vec3, scene.Rows)
	for i, _ := range data {
		data[i] = make([]geometry.Vec3, scene.Cols)
		peaks[i] = make([]geometry.Vec3, scene.Cols)
	}

	// Collect results
	var so_far time.Duration
	var highest, lowest geometry.Vec3
	highValue, lowValue := 0.0, math.Inf(+1)
	numPixels := scene.Rows * scene.Cols
	for i := 0; i < numPixels; i++ {
		// Print progress information every 500 pixels
		if i != 0 && i%500 == 0 {
			fmt.Printf("\rRendering %6.2f%%", 100*float64(i)/float64(scene.Rows*scene.Cols))
			so_far = time.Since(startTime)
			remaining := so_far * time.Duration(numPixels-i) / time.Duration(i)
			fmt.Printf(" (Time Remaining: %v at %0.1f pps)                \r", remaining, float64(i)/so_far.Seconds())
		}
		pixel := <-pixels

		if low := pixel.color.Abs(); low < lowValue {
			lowValue = low
			lowest = pixel.color
		}
		if high := pixel.color.Abs(); high > highValue {
			highValue = high
			highest = pixel.color
		}
		data[pixel.y][pixel.x] = pixel.color.CLAMPF()
		peaks[pixel.y][pixel.x] = pixel.color.PEAKS(0.8)
	}
	fmt.Println("\rRendering 100.00%")

	bloomed := BloomFilter(peaks, Config.BloomFactor)

	for y := 0; y < len(data); y++ {
		for x := 0; x < len(data[0]); x++ {
			c := data[y][x].Add(bloomed[y][x])
			c = CorrectColors(c).CLAMP()
			img.SetNRGBA(x, y, color.NRGBA{uint8(c.X), uint8(c.Y), uint8(c.Z), 255})
		}
	}
	clearLine()
	fmt.Println("\rDone!")
	fmt.Printf("Brightest pixel: %v intensity: %v\n", highest, highValue)
	fmt.Printf("Dimmest pixel: %v intensity: %v\n", lowest, lowValue)

	// Print duration
	fmt.Printf("Rendering took %v\n", time.Since(startTime))

	return img
}
