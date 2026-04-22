// ==================== Particle System ====================
class Particle {
    constructor(x, y, vx = null, vy = null) {
        this.x = x;
        this.y = y;
        this.vx = vx === null ? (Math.random() - 0.5) * 0.15 : vx;
        this.vy = vy === null ? (Math.random() - 0.5) * 0.15 : vy;
        this.radius = 2;
        this.mass = 1;
    }

    isAlive() {
        return true;
    }

    getAlpha() {
        return 1;
    }

    update(width, height) {
        this.x += this.vx;
        this.y += this.vy;

        // Boundary collision
        if (this.x - this.radius < 0 || this.x + this.radius > width) {
            this.vx *= -1;
            this.x = Math.max(this.radius, Math.min(width - this.radius, this.x));
        }
        if (this.y - this.radius < 0 || this.y + this.radius > height) {
            this.vy *= -1;
            this.y = Math.max(this.radius, Math.min(height - this.radius, this.y));
        }
    }

    draw(ctx) {
        ctx.beginPath();
        ctx.arc(this.x, this.y, this.radius, 0, Math.PI * 2);
        ctx.fillStyle = 'rgba(209, 213, 219, 0.5)';
        ctx.fill();
    }

    distanceTo(other) {
        const dx = this.x - other.x;
        const dy = this.y - other.y;
        return Math.sqrt(dx * dx + dy * dy);
    }
}

class ParticleSystem {
    constructor(canvas) {
        this.canvas = canvas;
        this.ctx = canvas.getContext('2d');
        this.particles = [];
        this.triangles = [];
        this.connectionDistance = 200;
        this.triangleDistance = 180;

        this.resize();
        this.initializeParticles();

        window.addEventListener('resize', () => this.resize());

        this.animate();
    }

    resize() {
        this.canvas.width = window.innerWidth;
        this.canvas.height = window.innerHeight;
    }

    initializeParticles() {
        const particleCount = Math.min(60, Math.max(30, Math.floor((this.canvas.width * this.canvas.height) / 40000)));
        for (let i = 0; i < particleCount; i++) {
            const x = Math.random() * this.canvas.width;
            const y = Math.random() * this.canvas.height;
            this.particles.push(new Particle(x, y));
        }
    }

    onMouseMove(e) {
        // Mouse interaction removed
    }

    findTriangles() {
        this.triangles = [];
        for (let i = 0; i < this.particles.length; i++) {
            for (let j = i + 1; j < this.particles.length; j++) {
                for (let k = j + 1; k < this.particles.length; k++) {
                    const p1 = this.particles[i];
                    const p2 = this.particles[j];
                    const p3 = this.particles[k];

                    const d12 = p1.distanceTo(p2);
                    const d23 = p2.distanceTo(p3);
                    const d31 = p3.distanceTo(p1);

                    if (d12 < this.triangleDistance && d23 < this.triangleDistance && d31 < this.triangleDistance) {
                        this.triangles.push([p1, p2, p3]);
                    }
                }
            }
        }
    }

    draw() {
        this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);

        // Draw triangles
        this.findTriangles();
        this.triangles.forEach(triangle => {
            const [p1, p2, p3] = triangle;
            this.ctx.beginPath();
            this.ctx.moveTo(p1.x, p1.y);
            this.ctx.lineTo(p2.x, p2.y);
            this.ctx.lineTo(p3.x, p3.y);
            this.ctx.closePath();
            this.ctx.strokeStyle = 'rgba(209, 213, 219, 0.15)';
            this.ctx.lineWidth = 1;
            this.ctx.stroke();
        });

        // Draw connections
        for (let i = 0; i < this.particles.length; i++) {
            for (let j = i + 1; j < this.particles.length; j++) {
                const d = this.particles[i].distanceTo(this.particles[j]);
                if (d < this.connectionDistance) {
                    const opacity = (1 - d / this.connectionDistance) * 0.25;
                    this.ctx.beginPath();
                    this.ctx.moveTo(this.particles[i].x, this.particles[i].y);
                    this.ctx.lineTo(this.particles[j].x, this.particles[j].y);
                    this.ctx.strokeStyle = `rgba(209, 213, 219, ${opacity})`;
                    this.ctx.lineWidth = 1;
                    this.ctx.stroke();
                }
            }
        }

        // Draw particles
        this.particles.forEach(p => p.draw(this.ctx));
    }

    animate() {
        this.particles.forEach(p => p.update(this.canvas.width, this.canvas.height));
        this.draw();
        requestAnimationFrame(() => this.animate());
    }
}

// ==================== Application State ====================
const appState = {
    currentMangaId: null,
    currentManga: null,
    selectedChapters: new Set(),
    downloadTasks: [],
    taskUpdateInterval: null,
    chaptersSort: { by: null, order: 'asc' },
};

// ==================== Toast Notifications ====================
function showToast(message, type = 'info') {
    const container = document.getElementById('toastContainer');
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.textContent = message;
    container.appendChild(toast);

    setTimeout(() => {
        toast.style.animation = 'slideInRight 0.4s ease reverse';
        setTimeout(() => toast.remove(), 400);
    }, 3000);
}

// ==================== API Functions ====================
async function queryManga(mangaId) {
    try {
        const response = await fetch(`/query/manga?mid=${mangaId}`);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        return await response.json();
    } catch (error) {
        console.error('Error querying manga:', error);
        throw error; // Re-throw to let caller handle
    }
}

async function downloadChapters(chapters) {
    try {
        const response = await fetch('/download/chapters', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(chapters),
        });
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        return await response.json();
    } catch (error) {
        console.error('Error submitting download:', error);
        showToast('Failed to submit download', 'error');
        return null;
    }
}

async function queryDownloadRecords() {
    try {
        const response = await fetch('/query/records');
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        return await response.json();
    } catch (error) {
        console.error('Error querying records:', error);
        return [];
    }
}

// ==================== Utility Functions ====================
function extractMangaIdFromUrl(url) {
    // Extract mid from URL like https://www.manhuagui.com/comic/<mid>/<cid>.html
    const match = url.match(/\/comic\/(\d+)\/(\d+)\.html/);
    return match ? { mid: parseInt(match[1]), cid: parseInt(match[2]) } : null;
}

function formatProgress(current, total) {
    if (total === 0) return '0%';
    return `${Math.round((current / total) * 100)}%`;
}


function showSearchLoading(show) {
    const loadingSpinner = document.getElementById('searchLoadingSpinner');
    const content = document.getElementById('searchResultsContent');
    if (show) {
        loadingSpinner.style.display = 'flex';
        content.classList.add('loading');
    } else {
        loadingSpinner.style.display = 'none';
        content.classList.remove('loading');
    }
}

function showSubmitLoading(show) {
    const loadingSpinner = document.getElementById('submitLoadingSpinner');
    const chaptersContainer = document.querySelector('.chapters-container');
    const submitButton = document.getElementById('submitDownloadButton');

    if (show) {
        loadingSpinner.style.display = 'flex';
        chaptersContainer.classList.add('loading');
        submitButton.disabled = true;
    } else {
        loadingSpinner.style.display = 'none';
        chaptersContainer.classList.remove('loading');
        submitButton.disabled = false;
    }
}

function toggleClearButton() {
    const input = document.getElementById('searchInput');
    const clearButton = document.getElementById('clearSearchButton');
    if (input.value.trim()) {
        clearButton.classList.remove('hidden');
    } else {
        clearButton.classList.add('hidden');
    }
}

function clearSearch() {
    const input = document.getElementById('searchInput');
    input.value = '';
    toggleClearButton();
    // Clear search results
    const mangaInfo = document.getElementById('mangaInfo');
    mangaInfo.classList.add('hidden');
    mangaInfo.innerHTML = '';
    document.getElementById('chaptersTableBody').innerHTML = '';
    appState.currentMangaId = null;
    appState.currentManga = null;
    appState.selectedChapters.clear();
}

// ==================== Search Results ====================
function displayMangaInfo(manga) {
    const mangaInfoDiv = document.getElementById('mangaInfo');
    mangaInfoDiv.innerHTML = `
        <img src="${manga.cover}" alt="${manga.name}" class="manga-cover">
        <div class="manga-details">
            <div class="manga-detail-item">
                <span class="manga-detail-label">Title:</span>
                <span class="manga-detail-value">${manga.name}</span>
            </div>
            <div class="manga-detail-item">
                <span class="manga-detail-label">Author:</span>
                <span class="manga-detail-value">${manga.author}</span>
            </div>
            <div class="manga-detail-item">
                <span class="manga-detail-label">Status:</span>
                <span class="manga-detail-value">${manga.status}</span>
            </div>
            <div class="manga-detail-item">
                <span class="manga-detail-label">Published:</span>
                <span class="manga-detail-value">${manga.date}</span>
            </div>
            <div class="manga-detail-item">
                <span class="manga-detail-label">Introduction:</span>
                <span class="manga-detail-value">${manga.introduction}</span>
            </div>
        </div>
    `;
    mangaInfoDiv.classList.remove('hidden');
}

function displayChapters() {
    if (!appState.currentManga) return;

    const tbody = document.getElementById('chaptersTableBody');
    tbody.innerHTML = '';
    appState.selectedChapters.clear();

    let chapters = [...appState.currentManga.contents];

    if (appState.chaptersSort.by) {
        chapters.sort((a, b) => {
            let valA, valB;
            if (appState.chaptersSort.by === 'title') {
                valA = a.title;
                valB = b.title;
            } else if (appState.chaptersSort.by === 'pages') {
                valA = parseInt(a.page);
                valB = parseInt(b.page);
            }
            if (appState.chaptersSort.order === 'asc') {
                return valA > valB ? 1 : valA < valB ? -1 : 0;
            } else {
                return valA < valB ? 1 : valA > valB ? -1 : 0;
            }
        });
    }

    chapters.forEach((chapter) => {
        const ids = extractMangaIdFromUrl(chapter.href);
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>
                <input type="checkbox" class="chapter-checkbox" 
                       ${ids ? '' : 'disabled'}>
            </td>
            <td>${chapter.title}</td>
            <td>${chapter.page}</td>
            <td><a href="${chapter.href}" target="_blank">View</a></td>
        `;

        const checkbox = row.querySelector('.chapter-checkbox');
        if (ids) {
            checkbox.dataset.mid = ids.mid;
            checkbox.dataset.cid = ids.cid;
            checkbox.addEventListener('change', updateSelectAllCheckbox);
        }

        tbody.appendChild(row);
    });

    updateSelectAllCheckbox();
}

function handleChapterSort(event) {
    const sortBy = event.target.dataset.sort;
    if (appState.chaptersSort.by === sortBy) {
        appState.chaptersSort.order = appState.chaptersSort.order === 'asc' ? 'desc' : 'asc';
    } else {
        appState.chaptersSort.by = sortBy;
        appState.chaptersSort.order = 'asc';
    }
    displayChapters();

    // Update sort icons
    document.querySelectorAll('.sort-icon').forEach(icon => icon.textContent = '');
    if (appState.chaptersSort.by) {
        const th = document.querySelector(`[data-sort="${appState.chaptersSort.by}"]`);
        th.querySelector('.sort-icon').textContent = appState.chaptersSort.order === 'asc' ? '↑' : '↓';
    }
}

function handleSearch() {
    const input = document.getElementById('searchInput');
    const mangaId = parseInt(input.value);

    if (!input.value || mangaId < 0) {
        showToast('Please enter a valid manga ID', 'info');
        return;
    }

    // Show loading spinner and hide content
    showSearchLoading(true);
    document.getElementById('searchButton').disabled = true;

    queryManga(mangaId).then(manga => {
        if (manga) {
            appState.currentMangaId = mangaId;
            appState.currentManga = manga; // Store current manga data
            displayMangaInfo(manga);
            displayChapters();
            showToast(`Loaded "${manga.name}"`, 'success');
        }
    }).catch(error => {
        console.error('Search error:', error);
        showToast('Failed to fetch manga data', 'error');
    }).finally(() => {
        // Hide spinner and enable button
        showSearchLoading(false);
        document.getElementById('searchButton').disabled = false;
    });
}

// ==================== Chapter Selection ====================
function updateSelectAllCheckbox() {
    const checkboxes = document.querySelectorAll('.chapter-checkbox:not(:disabled)');
    const selectAll = document.getElementById('selectAllCheckbox');
    const checkedCount = document.querySelectorAll('.chapter-checkbox:checked').length;

    selectAll.checked = checkedCount === checkboxes.length && checkboxes.length > 0;
    selectAll.indeterminate = checkedCount > 0 && checkedCount < checkboxes.length;

    // Update selected chapters set
    appState.selectedChapters.clear();
    document.querySelectorAll('.chapter-checkbox:checked').forEach(cb => {
        appState.selectedChapters.add({
            mid: parseInt(cb.dataset.mid),
            cid: parseInt(cb.dataset.cid),
        });
    });

    document.getElementById('submitDownloadButton').disabled = appState.selectedChapters.size === 0;
}

function toggleSelectAll() {
    const selectAll = document.getElementById('selectAllCheckbox');
    document.querySelectorAll('.chapter-checkbox:not(:disabled)').forEach(cb => {
        cb.checked = selectAll.checked;
    });
    updateSelectAllCheckbox();
}

async function submitDownload() {
    if (appState.selectedChapters.size === 0) {
        showToast('Please select at least one chapter', 'info');
        return;
    }

    const chapters = Array.from(appState.selectedChapters);
    showSubmitLoading(true);
    try {
        const result = await downloadChapters(chapters);

        if (result) {
            showToast(`Submitted ${chapters.length} chapter(s) for download`, 'success');
            updateDownloadTasks();
        }
    } finally {
        showSubmitLoading(false);
    }
}

// ==================== Download Tasks Management ====================
async function updateDownloadTasks() {
    const records = await queryDownloadRecords();
    appState.downloadTasks = records;
    renderDownloadTasks(records);
}

function renderDownloadTasks(records) {
    const ongoing = records.filter(r => !r.status.includes('success') && !r.status.includes('error'));
    const historical = records.filter(r => r.status.includes('success') || r.status.includes('error'));

    renderOngoingTasks(ongoing);
    renderHistoricalTasks(historical);

    // Show/hide bulk retry button
    const failedTasks = ongoing.filter(r => r.status.includes('error'));
    document.getElementById('bulkRetryButton').classList.toggle('hidden', failedTasks.length === 0);
}

function renderOngoingTasks(tasks) {
    const tbody = document.getElementById('ongoingTasksBody');
    tbody.innerHTML = '';

    tasks.forEach(task => {
        const isError = task.status.includes('error');
        const row = document.createElement('tr');
        
        const progressBar = task.status.includes('success') ? '' : `
            <div class="progress-bar">
                <div class="progress-fill" style="width: ${task.progress}%"></div>
            </div>
            <div class="progress-text">${task.progress.toFixed(1)}% (${task.count}/${task.total})</div>
        `;

        const statusClass = task.status.includes('success') ? 'success' : 
                          task.status.includes('error') ? 'error' :
                          task.status.includes('downloading') ? 'downloading' : 'waiting';

        const actionButton = isError ? 
            `<button class="retry-button" onclick="retryTask(${task.mid}, ${task.cid})">Retry</button>` :
            '';

        row.innerHTML = `
            <td><input type="checkbox" class="task-checkbox" ${isError ? '' : 'disabled'} 
                       data-mid="${task.mid}" data-cid="${task.cid}"></td>
            <td>${task.mname}</td>
            <td>${task.cname}</td>
            <td>${progressBar}</td>
            <td><span class="status-badge ${statusClass}">${task.status}</span></td>
            <td>${actionButton}</td>
        `;

        tbody.appendChild(row);
    });
}

function renderHistoricalTasks(tasks) {
    const tbody = document.getElementById('historicalTasksBody');
    tbody.innerHTML = '';

    tasks.forEach((task) => {
        const statusClass = task.status.includes('success') ? 'success' : 'error';
        const row = document.createElement('tr');
        
        row.innerHTML = `
            <td>${task.mname}</td>
            <td>${task.cname}</td>
            <td><span class="status-badge ${statusClass}">${task.status}</span></td>
            <td>${new Date().toLocaleString()}</td>
        `;

        tbody.appendChild(row);
    });
}

async function retryTask(mid, cid) {
    const result = await downloadChapters([{ mid, cid }]);
    if (result) {
        showToast('Retry submitted', 'success');
        updateDownloadTasks();
    }
}

function bulkRetry() {
    const checkboxes = document.querySelectorAll('.task-checkbox:checked');
    if (checkboxes.length === 0) {
        showToast('Please select failed tasks to retry', 'info');
        return;
    }

    const chapters = Array.from(checkboxes).map(cb => ({
        mid: parseInt(cb.dataset.mid),
        cid: parseInt(cb.dataset.cid),
    }));

    downloadChapters(chapters).then(result => {
        if (result) {
            showToast(`Resubmitted ${chapters.length} task(s)`, 'success');
            updateDownloadTasks();
        }
    });
}


// ==================== Event Listeners ====================
function setupEventListeners() {
    const searchInput = document.getElementById('searchInput');
    const searchButton = document.getElementById('searchButton');
    const clearSearchButton = document.getElementById('clearSearchButton');
    const selectAllCheckbox = document.getElementById('selectAllCheckbox');
    const submitDownloadButton = document.getElementById('submitDownloadButton');
    const bulkRetryButton = document.getElementById('bulkRetryButton');

    searchButton.addEventListener('click', handleSearch);
    searchInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') handleSearch();
    });
    searchInput.addEventListener('input', toggleClearButton);
    clearSearchButton.addEventListener('click', clearSearch);

    selectAllCheckbox.addEventListener('change', toggleSelectAll);
    submitDownloadButton.addEventListener('click', submitDownload);
    bulkRetryButton.addEventListener('click', bulkRetry);

    document.querySelectorAll('.sortable').forEach(th => {
        th.addEventListener('click', handleChapterSort);
    });
}

// ==================== Initialize ====================
document.addEventListener('DOMContentLoaded', () => {
    // Initialize particle system
    new ParticleSystem(document.getElementById('particleCanvas'));

    // Setup event listeners
    setupEventListeners();

    // Ensure spinners are hidden initially
    showSearchLoading(false);
    showSubmitLoading(false);

    // Initial task update and periodic refresh
    updateDownloadTasks();
    appState.taskUpdateInterval = setInterval(updateDownloadTasks, 3000);
});
