package desempenho

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// GetPerformanceData retorna os dados atuais de desempenho para uso externo (ex: Fiber)
func GetPerformanceData() PerformanceData {
	return performanceData
}

// Constante e struct permanecem as mesmas, mas agora pertencem a este pacote.
const historySize = 60

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

// performanceData é uma variável a nível de pacote, não exportada (letra minúscula),
// o que significa que só pode ser acessada por funções dentro deste pacote.
var performanceData PerformanceData

// updateMetrics coleta os dados e atualiza o histórico.
// Não precisa ser exportada pois só é chamada internamente por StartMonitoring.
func updateMetrics() {
	var lastReadBytes uint64 = 0
	var lastWriteBytes uint64 = 0
	var lastTime = time.Now()
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
		if err == nil {
			for _, stat := range ioStats {
				if stat.Name == "C:" || stat.Name == "PhysicalDrive0" {
					if lastReadBytes > 0 && lastWriteBytes > 0 && elapsed > 0 {
						readKBs = float64(stat.ReadBytes-lastReadBytes) / 1024.0 / elapsed
						writeKBs = float64(stat.WriteBytes-lastWriteBytes) / 1024.0 / elapsed
					}
					lastReadBytes = stat.ReadBytes
					lastWriteBytes = stat.WriteBytes
					break
				}
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

		if len(performanceData.CPUHistory) > historySize {
			performanceData.CPUHistory = performanceData.CPUHistory[1:]
			performanceData.MemHistory = performanceData.MemHistory[1:]
			performanceData.DiskHistory = performanceData.DiskHistory[1:]
		}
	}
}

// HandleAPI serve os dados em JSON. Exportada para uso externo.
func HandleAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(performanceData)
}

// handleRoot serve a página HTML. Não precisa ser exportada.
func handleRoot(w http.ResponseWriter, r *http.Request) {
	htmlContent := `
<!DOCTYPE html>
<html lang="pt-BR">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Dashboard de Desempenho (Go)</title>
	<style>
		body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; background-color: #121212; color: #e0e0e0; margin: 0; padding: 20px; }
		h1 { text-align: center; color: #00bcd4; }
		.dashboard { display: grid; grid-template-columns: repeat(auto-fit, minmax(400px, 1fr)); gap: 20px; margin-top: 20px; }
		.chart-container { background-color: #1e1e1e; padding: 20px; border-radius: 8px; box-shadow: 0 4px 8px rgba(0,0,0,0.3); }
	</style>
	<script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>
<body>
	<h1>Dashboard de Desempenho em Tempo Real</h1>
	<div class="dashboard">
		<div class="chart-container"><canvas id="cpuChart"></canvas></div>
		<div class="chart-container"><canvas id="memChart"></canvas></div>
		<div class="chart-container"><canvas id="diskChart"></canvas></div>
	</div>
	<script>
		function createChart(ctx, label, color) {
			return new Chart(ctx, {
				type: 'line',
				data: {
					labels: Array.from({ length: 60 }, (_, i) => i + 1),
					datasets: [{
						label: label, data: [], borderColor: color, backgroundColor: color + '33',
						borderWidth: 2, fill: true, tension: 0.4, pointRadius: 0,
					}]
				},
				options: {
					scales: {
						y: { beginAtZero: true, max: 100, ticks: { color: '#e0e0e0' } },
						x: { ticks: { color: '#e0e0e0' } }
					},
					plugins: { legend: { labels: { color: '#e0e0e0' } } }
				}
			});
		}
		const cpuChart = createChart(document.getElementById('cpuChart').getContext('2d'), 'Uso de CPU (%)', '#00bcd4');
		const memChart = createChart(document.getElementById('memChart').getContext('2d'), 'Uso de Memória (%)', '#ff9800');
		const diskChart = createChart(document.getElementById('diskChart').getContext('2d'), 'Uso de Disco (%)', '#4caf50');
		async function fetchDataAndUpdateCharts() {
			try {
				const response = await fetch('/api/performance');
				const data = await response.json();
				cpuChart.data.datasets[0].data = data.cpuHistory;
				cpuChart.update();
				memChart.data.datasets[0].data = data.memHistory;
				memChart.update();
				diskChart.data.datasets[0].data = data.diskHistory;
				diskChart.update();
			} catch (error) { console.error('Erro ao buscar dados:', error); }
		}
		setInterval(fetchDataAndUpdateCharts, 2000);
	</script>
</body>
</html>`
	fmt.Fprint(w, htmlContent)
}

// RegisterHandlersAndStartMonitoring é a ÚNICA função que precisa ser exportada (letra maiúscula).
// Ela configura tudo que este pacote precisa para funcionar.
func RegisterHandlersAndStartMonitoring() {
	// Inicializa os slices de histórico.
	performanceData.CPUHistory = make([]float64, historySize)
	performanceData.MemHistory = make([]float64, historySize)

	// Inicia a coleta de métricas em uma goroutine.
	go updateMetrics()

	// Configura os handlers para as URLs.
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/api/performance", HandleAPI)
}
