package api

import (
	"image"
	"math/rand"

	"github.com/MaxHalford/eaopt"
	"github.com/fogleman/gg"
	"github.com/rickyfitts/distributed-evolution/go/util"
)

type Shapes struct {
	Bounds  util.Vector
	Context *WorkerTask
	Members []Shape
	Type    string
}

// get the correct creation function based on the given shape type
func GetCreateShapeFunc(shapeType string) func(radius float64, bounds util.Vector, rng *rand.Rand) Shape {
	switch shapeType {
	case "triangles":
		return CreateTriangle
	case "polygons":
		return CreatePolygon
	default:
		return CreateCircle
	}
}

// returns a closure with a reference to the context that can be used to generate a random shapes object
func CreateShapesFactory(ctx *WorkerTask) func(rng *rand.Rand) eaopt.Genome {
	bounds := util.Vector{X: float64(ctx.TargetImage.Width), Y: float64(ctx.TargetImage.Height)}

	createShape := GetCreateShapeFunc(ctx.Task.Type)

	return func(rng *rand.Rand) eaopt.Genome {
		shapes := Shapes{
			Bounds:  bounds,
			Context: ctx,
			Members: make([]Shape, ctx.Task.Job.ShapesPerSlice),
			Type:    ctx.Task.Type,
		}

		for i := 0; i < int(ctx.Task.Job.ShapesPerSlice); i++ {
			shapes.Members[i] = createShape(float64(ctx.Task.Job.ShapeSize), bounds, rng)
		}

		return shapes
	}
}

func CreateShapesFactoryFromPopulation(ctx *WorkerTask, initialPopulation eaopt.Individuals) func(rng *rand.Rand) eaopt.Genome {
	population := initialPopulation

	return func(rng *rand.Rand) eaopt.Genome {
		index := len(population) - 1

		member := population[index]

		population = population[:index]

		s := member.Genome.(Shapes)

		s.Context = ctx

		return s
	}
}

func GetShapesFactory(ctx *WorkerTask, initialPopulation eaopt.Individuals) func(rng *rand.Rand) eaopt.Genome {
	if len(initialPopulation) > 0 {
		return CreateShapesFactoryFromPopulation(ctx, initialPopulation)
	}

	return CreateShapesFactory(ctx)
}

// draw the shapes to the given draw context
func (s Shapes) Draw(dc *gg.Context, offset util.Vector) {
	for _, m := range s.Members {
		m.Draw(dc, offset)
	}
}

// evaluates the fitness of the shapes instance
func (s Shapes) Evaluate() (float64, error) {
	// account for overdraw here
	overDraw := s.Context.Task.Job.OverDraw
	width := s.Context.TargetImage.Width + overDraw*2
	height := s.Context.TargetImage.Height + overDraw*2
	dc := gg.NewContext(width, height)

	s.Draw(dc, util.Vector{X: float64(overDraw), Y: float64(overDraw)})
	img := dc.Image()

	rect := image.Rect(overDraw, overDraw, width-overDraw, height-overDraw)
	out := util.GetSubImage(img, rect)

	// calculate fitness as the difference between the target and output images
	fitness := util.ImgDiff(out, s.Context.TargetImage.Image)

	s.Context.Mu.Lock()

	// if this is the best fit we've seen, save it
	if fitness > s.Context.BestFit.Fitness {
		s.Context.BestFit = Output{
			Fitness: fitness,
			Output:  img,
		}
	}

	s.Context.Mu.Unlock()

	return fitness, nil
}

// randomly replace members of the population with a new random shape
func (s Shapes) Mutate(rng *rand.Rand) {
	createShape := GetCreateShapeFunc(s.Type)

	for i := range s.Members {
		if rng.Float64() < s.Context.Task.Job.MutationRate {
			s.Members[i] = createShape(float64(s.Context.Task.Job.ShapeSize), s.Bounds, rng)
		}
	}
}

// randomly swap shapes between two populations
func (s Shapes) Crossover(g eaopt.Genome, rng *rand.Rand) {
	o := g.(Shapes)

	for i := range s.Members {
		if rng.Float64() < 0.5 {
			s.Members[i] = o.Members[i]
		} else {
			o.Members[i] = s.Members[i]
		}
	}
}

// create a new shapes instance with the same data
// copy all the data without pointers
func (s Shapes) Clone() eaopt.Genome {
	return Shapes{
		Bounds:  s.Bounds,
		Context: s.Context,
		Members: append([]Shape{}, s.Members...),
		Type:    s.Type,
	}
}

// creates a copy of the instance without context
func (s Shapes) CloneForSending() eaopt.Genome {
	return Shapes{
		Bounds:  s.Bounds,
		Members: append([]Shape{}, s.Members...),
		Type:    s.Type,
	}
}

func (s Shapes) At(i int) interface{} {
	return s.Members[i]
}

func (s Shapes) Set(i int, v interface{}) {
	s.Members[i] = v.(Shape)
}

func (s Shapes) Len() int {
	return len(s.Members)
}

func (s Shapes) Swap(i, j int) {
	s.Members[i], s.Members[j] = s.Members[j], s.Members[i]
}

func (s Shapes) Slice(a, b int) eaopt.Slice {
	slice := s.Clone()
	s.Members = s.Members[a:b]
	return slice.(eaopt.Slice)
}

func (s Shapes) Split(k int) (eaopt.Slice, eaopt.Slice) {
	s1 := s.Clone().(Shapes)
	s1.Members = s.Members[:k]

	s2 := s.Clone().(Shapes)
	s2.Members = s.Members[k:]

	return s1, s2
}

func (s Shapes) Append(q eaopt.Slice) eaopt.Slice {
	new := s.Clone().(Shapes)
	new.Members = append(s.Members, q.(Shapes).Members...)
	return new
}

func (s Shapes) Replace(q eaopt.Slice) {
	copy(s.Members, q.(Shapes).Members)
}

func (s Shapes) Copy() eaopt.Slice {
	return s.Clone().(Shapes)
}
