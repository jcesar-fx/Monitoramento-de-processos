async function fetchProcesses() {
    try {
        const res = await fetch('/api/processes');
        const processes = await res.json();

        const list = document.getElementById('process-list');
        list.innerHTML = ''; // clear current list

        processes.forEach(proc => {
            const li = document.createElement('li');
            li.className = "list-group-item";  // already sets basic styling

            li.innerHTML = `
            <div class="d-flex w-100 justify-content-between">
                <h5 class="mb-1">${proc.name}</h5>
                <small class="text-muted">PID(s): ${proc.pid.join(", ")}</small>
            </div>
            
            <p class="mb-1">
                <span class="badge bg-primary me-1">CPU: ${proc.cpuusage.toFixed(2)}%</span>
                <span class="badge bg-success">Memory: ${proc.memoryusage.toFixed(2)}%</span>
            </p>

            <small class="text-muted">Started: ${new Date(proc.created).toLocaleString()}</small><br>
            <small class="text-break text-secondary">Path: ${proc.path}</small>
            `;
            list.appendChild(li);
        });
    } catch (error) {
        console.error('Failed to fetch processes:', error);
    }
}

// fetch initially
fetchProcesses();

// fetch every 5 seconds
setInterval(fetchProcesses, 2000);