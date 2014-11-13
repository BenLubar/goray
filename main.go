package main

import (
	"flag"
	"fmt"
	"github.com/BenLubar/goray/geometry"
	"github.com/BenLubar/goray/render"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
)

var (
	input    = new(string) //flag.String("in", "default", "The file describing the scene")
	cores    = flag.Int("cores", 2, "The number of cores to use on the machine")
	chunks   = flag.Int("chunks", 8, "The number of chunks to use for parallelism")
	fov      = flag.Int("fov", 75, "The field of view of the rendered image")
	cols     = flag.Int("w", 800, "The width in pixels of the rendered image")
	rows     = flag.Int("h", 600, "The height in pixels of the rendered image")
	seed     = flag.Int64("seed", 1, "The seed for the random number generator")
	output   = flag.String("out", "out%04d.png", "Output file for the rendered scene")
	bloom    = flag.Int("bloom", 10, "The number of iteration to run the bloom filter")
	mindepth = flag.Int("depth", 2, "The minimum recursion depth used for the rays")
	rays     = flag.Int("rays", 10, "The number of rays used to sample each pixel")
	caustics = flag.Int("caustics", -1, "The depth of the caustic photon tracing before the render")
	gamma    = flag.Float64("gamma", 2.2, "The factor to use for gamma correction")

	skipTop    = flag.Int("skiptop", 0, "The number of pixels to skip calculating starting from the top of the image")
	skipLeft   = flag.Int("skipleft", 0, "The number of pixels to skip calculating starting from the left side of the image")
	skipRight  = flag.Int("skipright", 0, "The number of pixels to skip calculating starting from the right side of the image")
	skipBottom = flag.Int("skipbottom", 0, "The number of pixels to skip calculating starting from the bottom of the image")

	// Profiling information
	cpuprofile = flag.String("cpuprofile", "", "Write cpu profile informaion to file")
	memprofile = flag.String("memprofile", "", "Write memory profile informaion to file")
)

func main() {
	flag.Parse()

	rand.Seed(*seed)

	render.Config.NumRays = *rays
	render.Config.Caustics = *caustics
	render.Config.BloomFactor = *bloom
	render.Config.MinDepth = *mindepth
	render.Config.GammaFactor = *gamma

	render.Config.Skip.Top = *skipTop
	render.Config.Skip.Left = *skipLeft
	render.Config.Skip.Right = *skipRight
	render.Config.Skip.Bottom = *skipBottom

	wantedCPUs := *cores
	if wantedCPUs < 1 {
		wantedCPUs = 1
	}
	fmt.Printf("Running on %v/%v CPU cores\n", wantedCPUs, runtime.NumCPU())
	runtime.GOMAXPROCS(wantedCPUs)

	if wantedCPUs > *chunks {
		log.Print("Warning: chunks setting is lower than the number of cores - not all cores will be used")
	}

	if *rows%*chunks != 0 {
		log.Fatal("The images height needs to be evenly divisible by chunks")
	}

	render.Config.Chunks = *chunks

	if *cpuprofile != "" {
		cpupf, err := os.Create(*cpuprofile)
		fmt.Println("Writing CPU profiling information to file:", *cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(cpupf)
		defer pprof.StopCPUProfile()
	}

	if *memprofile != "" {
		fmt.Println("Writing Memory profiling information to file:", *memprofile)
	} else {
		runtime.MemProfileRate = 0
	}

	fmt.Printf("Rendering %vx%v sized image with %v rays per pixel to %v\n", *cols, *rows, *rays, *output)

	// "Real world" frustrum
	height := 2.0
	width := height * float64(*cols) / float64(*rows) // Aspect ratio?
	angle := math.Pi * float64(*fov) / 180.0

	scene := geometry.ParseScene(*input, width, height, angle, *cols, *rows)

	const x_shift = 5
	const fps = 60
	for i := 0; i <= 2*x_shift*fps; i++ {
		scene.Camera.X = -(float64(i)/fps - x_shift)

		img := render.Render(scene)

		file, err := os.Create(fmt.Sprintf(*output, i))
		if err != nil {
			log.Fatal(err)
		}

		if err = png.Encode(file, img); err != nil {
			log.Fatal(err)
		}

		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}

	if *memprofile != "" {
		mempf, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(mempf)
		defer mempf.Close()
	}
}
