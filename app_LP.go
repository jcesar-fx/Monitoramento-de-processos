package main

import (
	"log"
	"runtime"

	"app_LP/desempenho"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/django/v3"

	"github.com/shirou/gopsutil/v3/process"
)

var numCPUs int = runtime.NumCPU()

func sameNamePIDs() [][]int {
	listaDeProcessos, _ := process.Processes()

	sameNamePIDs := [][]int{}
	nomesVistos := []string{}

	for i := 0; i < len(listaDeProcessos); i++ {
		nome, _ := listaDeProcessos[i].Name()
		visto := false

		for k := 0; k < len(nomesVistos); k++ {
			if nome == nomesVistos[k] {
				visto = true
			}
		}
		if visto == false {

			nomesVistos = append(nomesVistos, nome)
			if nome > "" {
				//log.Printf("Current process name: %s \n", nome)

				equalProcess := []int{int(listaDeProcessos[i].Pid)}
				for j := 0; j < len(listaDeProcessos); j++ {
					nomeComparativo, _ := listaDeProcessos[j].Name()

					if nome == nomeComparativo {
						equalProcess = append(equalProcess, int(listaDeProcessos[j].Pid))
					}
				}
				sameNamePIDs = append(sameNamePIDs, equalProcess)
			}
		}
	}
	return sameNamePIDs
}

type detailedProcess struct {
	Pid         []int   `json:"pid"`
	Name        string  `json:"name"`
	Cpuusage    float64 `json:"cpuusage"`
	Created     int64   `json:"created"`
	Path        string  `json:"path"`
	Memoryusage float32 `json:"memoryusage"`
}

func detailProcesses() []detailedProcess {
	PIDs := sameNamePIDs()
	dProcSlice := []detailedProcess{}

	for i := 0; i < len(PIDs); i++ {

		grupoPIDs := PIDs[i]
		var dProc detailedProcess
		dProc.Pid = grupoPIDs

		for j := 0; j < len(grupoPIDs); j++ {
			proc, _ := process.NewProcess(int32(grupoPIDs[j]))

			if j == 0 {
				dProc.Name, _ = proc.Name()
				dProc.Created, _ = proc.CreateTime()
				dProc.Path, _ = proc.Exe()
			}

			tempCPUusage, _ := proc.CPUPercent()
			tempMemUsage, _ := proc.MemoryPercent()

			dProc.Cpuusage += tempCPUusage / float64(numCPUs)
			dProc.Memoryusage += tempMemUsage / float32(numCPUs)

		}
		dProcSlice = append(dProcSlice, dProc)
	}
	return dProcSlice
}

func main() {
	engine := django.New("./views", ".django")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	// Inicia o monitoramento de desempenho em background
	desempenho.RegisterHandlersAndStartMonitoring()

	app.Get("/desempenho", func(c *fiber.Ctx) error {

		return c.Render("desempenho", fiber.Map{
			"Title": "Parte de Desempenho",
		})
	})

	app.Get("/about", func(c *fiber.Ctx) error {

		return c.Render("about", fiber.Map{
			"Title": "Sobre o projeto",
		})
	})
	app.Static("/static", "./views/static")

	// API de desempenho integrada ao Fiber
	app.Get("/api/performance", func(c *fiber.Ctx) error {
		data := desempenho.GetPerformanceData()
		return c.JSON(data)
	})

	app.Get("/api/processes", func(c *fiber.Ctx) error {

		processes := detailProcesses()
		return c.JSON(processes)
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("processos", fiber.Map{})
	})

	log.Fatal(app.Listen(":3000"))
}
