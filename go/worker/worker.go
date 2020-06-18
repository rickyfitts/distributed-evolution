package worker

import (
	"fmt"
	"image"
	"log"
	"time"

	"github.com/MaxHalford/eaopt"

	"github.com/rickyfitts/distributed-evolution/api"
	"github.com/rickyfitts/distributed-evolution/util"
)

type Image struct {
	Image  image.Image
	Width  int
	Height int
}

type Worker struct {
	MutationRate float64
	CurrentTask  api.Task
	TargetImage  Image

	ga *eaopt.GA
}

func createGA() *eaopt.GA {
	ga, err := eaopt.NewDefaultGAConfig().NewGA()
	if err != nil {
		log.Fatal("error creating ga: ", err)
	}

	return ga
}

// TODO figure out how we can set an initial population to start from
// maybe make another version of createTriangleFactory that accepts a seed population
func (w *Worker) RunTask(task api.Task) {
	util.DPrintf("assigned task %v\n", task.ID)

	util.DPrintf("decoding image...")
	img := util.DecodeImage(task.TargetImage)
	width, height := util.GetImageDimensions(img)

	util.DPrintf("saving task data...")
	w.CurrentTask = task
	w.TargetImage = Image{
		Image:  img,
		Width:  width,
		Height: height,
	}

	util.DPrintf("preparing ga...")

	w.ga.NGenerations = 100
	w.ga.PopSize = 10

	w.ga.Callback = func(ga *eaopt.GA) {
		util.DPrintf("best fitness at generation %d: %f\n", ga.Generations, ga.HallOfFame[0].Fitness)

		task.Generation = ga.Generations
		task.Population = make(eaopt.Individuals, len(ga.Populations[0].Individuals))

		// make a copy of each individual without the context pointer to the worker state
		for i, indv := range ga.Populations[0].Individuals {
			copy := indv
			genome := copy.Genome.Clone()

			if task.Type == "triangles" {
				t := genome.(Triangle)
				t.Context = nil
				genome = t
			}

			copy.Genome = genome

			task.Population[i] = copy
		}

		util.DPrintf("updating master")

		api.Update(task)
	}

	Factory := createTriangleFactory(w)

	util.DPrintf("evolving...")

	err := w.ga.Minimize(Factory)
	if err != nil {
		fmt.Println(err)
	}
}

func Run() {
	w := Worker{MutationRate: 0.8}
	w.ga = createGA()

	// wait for master to initialize
	time.Sleep(10 * time.Second)

	for {
		// TODO handle errors by waiting and trying again
		task := api.GetTask()

		// if generation is zero this is an empty response, if so just wait for more work
		if task.Generation != 0 {
			w.RunTask(task)
		}

		time.Sleep(time.Second)
	}
}
