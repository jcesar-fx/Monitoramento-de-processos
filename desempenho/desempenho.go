package desempenho

import (
	"encoding/json" // "encoding/json" // Importa o pacote de JSON para codificação e decodificação
	"log"           // Ajuda a registrar as mensagens de Log para debugar
	"net/http"      // Pacote HTTP
	"time"          // Controla o tempo de execução

	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// GetPerformanceData retorna os dados atuais de desempenho para uso extern
func GetPerformanceData() PerformanceData {
	return performanceData
}

const historySize = 60

// PerformanceData contém os dados de desempenho do sistema.
type PerformanceData struct {
	CPUHistory   []float64 `json:"cpuHistory"`
	MemHistory   []float64 `json:"memHistory"`
	DiskHistory  []float64 `json:"diskHistory"`
	CPUFreq      float64   `json:"cpuFreq"` // em MHz
	MemUsedMB    float64   `json:"memUsedMB"`
	MemTotalMB   float64   `json:"memTotalMB"`
	DiskUsedGB   float64   `json:"diskUsedGB"`
	DiskTotalGB  float64   `json:"diskTotalGB"`
	DiskReadKBs  float64   `json:"diskReadKBs"`
	DiskWriteKBs float64   `json:"diskWriteKBs"`
}

// performanceData é uma variável a nível de pacote
// o que significa que só pode ser acessada por funções dentro deste pacote.
var performanceData PerformanceData

// updateMetrics coleta os dados e atualiza o histórico.
// Não precisa ser exportada pois só é chamada internamente por StartMonitoring.
var lastReadBytes uint64 = 0
var lastWriteBytes uint64 = 0
var lastTime = time.Now()

func updateMetrics() {
	for {
		cpuPerc, err := cpu.Percent(time.Second, false)
		if err != nil {
			log.Printf("Erro ao obter CPU: %v", err)
			cpuPerc = []float64{0}
		}

		memInfo, err := mem.VirtualMemory()
		if err != nil {
			log.Printf("Erro ao obter Memória: %v", err)
			memInfo = &mem.VirtualMemoryStat{UsedPercent: 0}
		}

		diskInfo, err := disk.Usage("C:\\")
		if err != nil {
			log.Printf("Erro ao obter Disco: %v", err)
			diskInfo = &disk.UsageStat{UsedPercent: 0}
		}

		// Frequência da CPU (MHz)
		cpuFreqs, err := cpu.Info()
		var freq float64 = 0
		if err == nil && len(cpuFreqs) > 0 {
			freq = cpuFreqs[0].Mhz
		}

		// Memória usada e total (MB)
		memUsedMB := float64(memInfo.Used) / 1024.0 / 1024.0
		memTotalMB := float64(memInfo.Total) / 1024.0 / 1024.0

		// Disco usado e total (GB)
		diskUsedGB := float64(diskInfo.Used) / 1024.0 / 1024.0 / 1024.0
		diskTotalGB := float64(diskInfo.Total) / 1024.0 / 1024.0 / 1024.0

		// Velocidade de leitura/gravação do disco (KB/s)
		ioStats, err := disk.IOCounters()
		var readKBs, writeKBs float64
		now := time.Now()
		elapsed := now.Sub(lastTime).Seconds()
		if err == nil && elapsed > 0 {
			if stat, ok := ioStats["C:"]; ok {
				readKBs = float64(stat.ReadBytes-lastReadBytes) / 1024.0 / elapsed
				writeKBs = float64(stat.WriteBytes-lastWriteBytes) / 1024.0 / elapsed
				lastReadBytes = stat.ReadBytes
				lastWriteBytes = stat.WriteBytes
			}
		}
		lastTime = now

		performanceData.CPUHistory = append(performanceData.CPUHistory, cpuPerc[0])
		performanceData.MemHistory = append(performanceData.MemHistory, memInfo.UsedPercent)
		// Calcula o percentual de uso do disco manualmente usando os valores em GB
		var diskPercent float64 = 0
		if diskTotalGB > 0 {
			diskPercent = (diskUsedGB / diskTotalGB) * 100
		}
		performanceData.DiskHistory = append(performanceData.DiskHistory, diskPercent)
		performanceData.CPUFreq = freq
		performanceData.MemUsedMB = memUsedMB
		performanceData.MemTotalMB = memTotalMB
		performanceData.DiskUsedGB = diskUsedGB
		performanceData.DiskTotalGB = diskTotalGB
		performanceData.DiskReadKBs = readKBs
		performanceData.DiskWriteKBs = writeKBs
                // Cria uma lista para guardar os 60 "tiques" do histórico de métricas
		if len(performanceData.CPUHistory) > historySize {
			performanceData.CPUHistory = performanceData.CPUHistory[1:]
			performanceData.MemHistory = performanceData.MemHistory[1:]
			performanceData.DiskHistory = performanceData.DiskHistory[1:]
		}
		time.Sleep(1 * time.Second) // Atualiza a cada segundo
	}
}

// HandleAPI sai como JSON. Exportada para uso externo.
func HandleAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(performanceData)
}

// RegisterHandlersAndStartMonitoring é a ÚNICA função que precisa ser exportada
// Ela configura tudo que este pacote precisa para funcionar.
func RegisterHandlersAndStartMonitoring() {

	performanceData.CPUHistory = make([]float64, historySize)
	performanceData.MemHistory = make([]float64, historySize)

	// Inicia a coleta de métricas em uma goroutine.
	go updateMetrics()

	// Configura os handlers para as URLs.
	http.HandleFunc("/api/performance", HandleAPI)
}
