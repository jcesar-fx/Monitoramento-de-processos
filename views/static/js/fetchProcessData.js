let sortField = 'name';
let ascending = true;


document.querySelectorAll('.dropdown-item').forEach(item => {
  item.addEventListener('click', function () {
    sortField = this.dataset.sort;
    document.getElementById('sortDropdown').textContent = `Ordenar por: ${this.textContent}`;
    fetchProcesses();
  });
});


document.getElementById('sort-toggle').onclick = () => {
  ascending = !ascending;
  document.getElementById('sort-toggle').innerText = ascending ? "Crescente ↑" : "Decrescente ↓";
  fetchProcesses(); 
};

async function fetchProcesses() {
  try {
    const res = await fetch('/api/processes');
    const processes = await res.json();

    processes.sort((a, b) => {
      let valA, valB;
      switch (sortField) {
        case 'name':
          valA = a.name.toLowerCase();
          valB = b.name.toLowerCase();
          break;
        case 'cpu':
          valA = a.cpuusage;
          valB = b.cpuusage;
          break;
        case 'mem':
          valA = a.memoryusage;
          valB = b.memoryusage;
          break;
        case 'pid':
          valA = a.pid.length;
          valB = b.pid.length;
          break;
      }

      if (valA < valB) return ascending ? -1 : 1;
      if (valA > valB) return ascending ? 1 : -1;
      return 0;
    });

    const list = document.getElementById('process-list');
    list.innerHTML = '';

    processes.forEach(proc => {
      const li = document.createElement('li');
      li.className = "list-group-item";

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

fetchProcesses();
setInterval(fetchProcesses, 2000);
