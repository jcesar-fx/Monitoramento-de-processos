let cpuChart, memChart, diskChart;
function createChart(ctx, label, color) {
    return new Chart(ctx, {
        type: 'line',
        data: {
            
            labels: Array.from({ length: 60 }, (_, i) => i + 1),
            datasets: [{
                label: label, data: [], borderColor: color, backgroundColor: color + '33',
                borderWidth: 2, fill: true, tension: 0.0, pointRadius: 0,
            }]
        },
        options: {
            animations: false,
            scales: {
                y: { beginAtZero: true, max: 100 },
                x: {}
            },
            plugins: { legend: { labels: { color: '#222' } } }
        }
    });
}
window.onload = function() {
    cpuChart = createChart(document.getElementById('cpuChart').getContext('2d'), 'CPU (%)', '#00bcd4');
    memChart = createChart(document.getElementById('memChart').getContext('2d'), 'Memória (%)', '#ff9800');
    diskChart = createChart(document.getElementById('diskChart').getContext('2d'), 'Disco (%)', '#4caf50');
    fetchAndUpdate();
    setInterval(fetchAndUpdate, 1000);
}
async function fetchAndUpdate() {
    try {
        const resp = await fetch('/api/performance');
        const data = await resp.json();
        document.getElementById('cpu').textContent = data.cpuHistory[data.cpuHistory.length-1]?.toFixed(1) || '0';
        document.getElementById('cpusidebar').textContent = data.cpuHistory[data.cpuHistory.length-1]?.toFixed(1) || '0';
        document.getElementById('mem').textContent = data.memHistory[data.memHistory.length-1]?.toFixed(1) || '0';
        document.getElementById('memsidebar').textContent = data.memHistory[data.memHistory.length-1]?.toFixed(1) || '0';
        // Não exibe mais o percentual do disco, apenas armazenamento e velocidades
        // Exibir valores absolutos
        document.getElementById('cpuabs').textContent = data.cpuFreq ? `(${(data.cpuFreq/1000).toFixed(2)} GHz)` : '';
        document.getElementById('memabs').textContent = (data.memUsedMB && data.memTotalMB) ? `(${(data.memUsedMB/1024).toFixed(2)} GB / ${(data.memTotalMB/1024).toFixed(2)} GB)` : '';
        document.getElementById('diskabs').textContent = (data.diskUsedGB && data.diskTotalGB) ? `(${data.diskUsedGB.toFixed(1)} GB / ${data.diskTotalGB.toFixed(1)} GB)` : '';
        document.getElementById('diskvel').textContent = (data.diskReadKBs || data.diskWriteKBs) ? `[Leitura: ${data.diskReadKBs?.toFixed(1) || 0} KB/s | Gravação: ${data.diskWriteKBs?.toFixed(1) || 0} KB/s]` : '';
        cpuChart.data.datasets[0].data = data.cpuHistory;
        cpuChart.update();
        memChart.data.datasets[0].data = data.memHistory;
        memChart.update();
        diskChart.data.datasets[0].data = data.diskHistory;
        diskChart.update();
    } catch (e) {
        console.error('Erro ao buscar desempenho:', e);
    }
}