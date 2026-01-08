// API Base URL
const API_BASE = '/api/v1';

// State
let currentFiles = [];
let allImages = [];

// Tab Switching
document.querySelectorAll('.tab-btn').forEach(btn => {
    btn.addEventListener('click', () => {
        const tabName = btn.dataset.tab;
        switchTab(tabName);
    });
});

function switchTab(tabName) {
    // Update buttons
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');

    // Update content
    document.querySelectorAll('.tab-pane').forEach(pane => {
        pane.classList.remove('active');
    });
    document.getElementById(tabName).classList.add('active');

    // Load data for tab
    if (tabName === 'warehouse') {
        loadWarehouse();
    }
}

// ============ UPLOAD TAB ============

// Drag and Drop
const dropZone = document.getElementById('dropZone');
const fileInput = document.getElementById('fileInput');
const folderInput = document.getElementById('folderInput');

// Button handlers (with stopPropagation to prevent dropZone click)
document.getElementById('chooseFilesBtn').addEventListener('click', (e) => {
    e.stopPropagation();
    fileInput.click();
});

document.getElementById('chooseFolderBtn').addEventListener('click', (e) => {
    e.stopPropagation();
    folderInput.click();
});

// Drag and drop
dropZone.addEventListener('dragover', (e) => {
    e.preventDefault();
    dropZone.classList.add('drag-over');
});

dropZone.addEventListener('dragleave', () => {
    dropZone.classList.remove('drag-over');
});

dropZone.addEventListener('drop', (e) => {
    e.preventDefault();
    dropZone.classList.remove('drag-over');

    // Try to get files from dataTransfer
    const items = e.dataTransfer.items;
    const files = [];

    if (items) {
        // Use DataTransferItemList interface
        for (let i = 0; i < items.length; i++) {
            if (items[i].kind === 'file') {
                const file = items[i].getAsFile();
                if (file && file.type.startsWith('image/')) {
                    files.push(file);
                }
            }
        }
    } else {
        // Fallback to DataTransfer.files
        const droppedFiles = Array.from(e.dataTransfer.files);
        files.push(...droppedFiles.filter(f => f.type.startsWith('image/')));
    }

    if (files.length > 0) {
        handleFiles(files);
    } else {
        alert('No image files found. Please drop image files (.jpg, .png, .gif, .webp)');
    }
});

// File input change
fileInput.addEventListener('change', (e) => {
    const files = Array.from(e.target.files);
    handleFiles(files);
    fileInput.value = ''; // Reset so same files can be selected again
});

// Folder input change
folderInput.addEventListener('change', (e) => {
    const files = Array.from(e.target.files).filter(f => f.type.startsWith('image/'));
    if (files.length > 0) {
        console.log(`Selected ${files.length} images from folder(s)`);
        handleFiles(files);
    } else {
        alert('No images found in the selected folder(s)');
    }
    folderInput.value = ''; // Reset
});

function handleFiles(files) {
    currentFiles = files;
    displayUploadQueue(files);
}

function displayUploadQueue(files) {
    const queue = document.getElementById('uploadQueue');
    if (files.length === 0) {
        queue.innerHTML = '';
        return;
    }

    queue.innerHTML = `
        <h3>Ready to Upload (${files.length} file${files.length > 1 ? 's' : ''})</h3>
        ${files.map((f, i) => `
            <div class="upload-item" data-index="${i}">
                <span class="upload-item-name">${f.name}</span>
                <span class="upload-item-status" id="status-${i}">⏳ Ready</span>
            </div>
        `).join('')}
        <button class="btn" onclick="uploadAll()" style="margin-top: 15px;">Upload All</button>
    `;
}

async function uploadAll() {
    if (currentFiles.length === 0) {
        alert('No files selected');
        return;
    }

    const title = document.getElementById('uploadTitle').value || '';
    const artist = document.getElementById('uploadArtist').value || 'Unknown';
    const tagsInput = document.getElementById('uploadTags').value;
    const tags = tagsInput ? tagsInput.split(',').map(t => t.trim()) : [];

    for (let i = 0; i < currentFiles.length; i++) {
        const file = currentFiles[i];
        const statusEl = document.getElementById(`status-${i}`);

        try {
            statusEl.textContent = '⏳ Uploading...';
            statusEl.className = 'upload-item-status status-processing';

            const fileTitle = title || file.name.replace(/\.[^/.]+$/, '');
            const result = await uploadImage(file, fileTitle, artist, tags);

            statusEl.textContent = `✅ Uploaded (ID: ${result.id.substring(0, 8)})`;
            statusEl.className = 'upload-item-status status-success';
        } catch (err) {
            statusEl.textContent = `❌ Failed: ${err.message}`;
            statusEl.className = 'upload-item-status status-error';
        }
    }

    showStatus('Upload complete! Processing images with AI...', 'success');

    // Clear form after 3 seconds
    setTimeout(() => {
        currentFiles = [];
        document.getElementById('uploadQueue').innerHTML = '';
        fileInput.value = '';
        document.getElementById('uploadTitle').value = '';
        document.getElementById('uploadTags').value = '';
    }, 3000);
}

async function uploadImage(file, title, artist, tags) {
    const formData = new FormData();
    formData.append('image', file);
    formData.append('title', title);
    formData.append('artist', artist);
    formData.append('tags', JSON.stringify(tags));

    const response = await fetch(`${API_BASE}/images/upload`, {
        method: 'POST',
        body: formData
    });

    if (!response.ok) {
        const error = await response.text();
        throw new Error(error);
    }

    return await response.json();
}

function showStatus(message, type) {
    const statusDiv = document.getElementById('uploadStatus');
    const msgDiv = document.createElement('div');
    msgDiv.className = `status-message ${type}`;
    msgDiv.textContent = message;
    statusDiv.appendChild(msgDiv);

    setTimeout(() => msgDiv.remove(), 5000);
}

// ============ WAREHOUSE TAB ============

async function loadWarehouse() {
    const grid = document.getElementById('warehouseGrid');
    grid.innerHTML = '<div class="loading">Loading images...</div>';

    try {
        const response = await fetch(`${API_BASE}/images`);
        if (!response.ok) throw new Error('Failed to load images');

        const data = await response.json();
        allImages = data.images || [];

        // Update count
        document.getElementById('imageCount').textContent = `${allImages.length} image${allImages.length !== 1 ? 's' : ''}`;

        // Extract unique categories
        const categories = [...new Set(allImages.map(img => img.category).filter(Boolean))];
        updateCategoryFilter(categories);

        displayImages(allImages);
    } catch (err) {
        grid.innerHTML = `<div class="empty-state">
            <div class="empty-state-icon">❌</div>
            <p>Failed to load images: ${err.message}</p>
        </div>`;
    }
}

function updateCategoryFilter(categories) {
    const select = document.getElementById('categoryFilter');
    select.innerHTML = '<option value="">All Categories</option>';
    categories.forEach(cat => {
        const option = document.createElement('option');
        option.value = cat;
        option.textContent = cat;
        select.appendChild(option);
    });

    select.onchange = () => {
        const selected = select.value;
        const filtered = selected ? allImages.filter(img => img.category === selected) : allImages;
        displayImages(filtered);
    };
}

function displayImages(images) {
    const grid = document.getElementById('warehouseGrid');

    if (images.length === 0) {
        grid.innerHTML = `<div class="empty-state">
            <div class="empty-state-icon">📭</div>
            <p>No images found</p>
        </div>`;
        return;
    }

    grid.innerHTML = images.map(img => `
        <div class="image-card" onclick="showImageModal('${img.id}')">
            <img src="/data/${img.thumbnail_path || img.file_path}" alt="${img.title}"
                 onerror="this.src='data:image/svg+xml,<svg xmlns=%22http://www.w3.org/2000/svg%22 width=%22200%22 height=%22200%22><rect fill=%22%23ddd%22 width=%22200%22 height=%22200%22/><text x=%2250%%22 y=%2250%%22 text-anchor=%22middle%22 dy=%22.3em%22 fill=%22%23999%22>No Image</text></svg>'">
            <div class="image-card-body">
                <div class="image-card-title">${img.title || 'Untitled'}</div>
                <div class="image-card-category">${img.category || 'uncategorized'}</div>
                ${img.tags && img.tags.length > 0 ? `
                    <div class="image-card-tags">
                        ${img.tags.slice(0, 3).map(tag => `<span class="tag">${tag}</span>`).join('')}
                    </div>
                ` : ''}
            </div>
        </div>
    `).join('');
}

async function showImageModal(imageId) {
    const modal = document.getElementById('imageModal');
    const modalBody = document.getElementById('modalBody');

    modal.classList.add('show');
    modalBody.innerHTML = '<div class="loading">Loading...</div>';

    try {
        const response = await fetch(`${API_BASE}/images/${imageId}`);
        if (!response.ok) throw new Error('Image not found');

        const img = await response.json();

        modalBody.innerHTML = `
            <img src="/data/${img.file_path || img.FilePath}" alt="${img.title || img.Title}">
            <h2>${img.title || img.Title || 'Untitled'}</h2>
            <p><strong>Artist:</strong> ${img.artist || img.Artist || 'Unknown'}</p>
            <p><strong>Category:</strong> ${img.category || img.Category || 'uncategorized'}</p>
            <p><strong>Uploaded:</strong> ${img.uploaded_at || img.UploadedAt || 'N/A'}</p>
            ${img.description || img.AIAnalysis?.Description ? `
                <p><strong>Description:</strong> ${img.description || img.AIAnalysis.Description}</p>
            ` : ''}
            ${img.tags && img.tags.length > 0 ? `
                <p><strong>Tags:</strong> ${img.tags.join(', ')}</p>
            ` : ''}
        `;
    } catch (err) {
        modalBody.innerHTML = `<p>Error loading image: ${err.message}</p>`;
    }
}

function closeModal() {
    document.getElementById('imageModal').classList.remove('show');
}

// Close modal on outside click
document.getElementById('imageModal').addEventListener('click', (e) => {
    if (e.target.id === 'imageModal') {
        closeModal();
    }
});

// ============ SEARCH TAB ============

async function performSearch() {
    const query = document.getElementById('searchInput').value.trim();
    if (!query) {
        alert('Please enter a search query');
        return;
    }

    const resultsDiv = document.getElementById('searchResults');
    resultsDiv.innerHTML = '<div class="loading">Searching...</div>';

    try {
        const response = await fetch(`${API_BASE}/search`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ query, limit: 10 })
        });

        if (!response.ok) throw new Error('Search failed');

        const data = await response.json();
        displaySearchResults(data, query);
    } catch (err) {
        resultsDiv.innerHTML = `<div class="empty-state">
            <div class="empty-state-icon">❌</div>
            <p>Search failed: ${err.message}</p>
        </div>`;
    }
}

function displaySearchResults(data, query) {
    const resultsDiv = document.getElementById('searchResults');

    if (!data.results || data.results.length === 0) {
        resultsDiv.innerHTML = `<div class="empty-state">
            <div class="empty-state-icon">🔍</div>
            <p>No results found for "${query}"</p>
            <p style="color: #999; font-size: 0.9em;">Try different keywords or wait for images to finish processing</p>
        </div>`;
        return;
    }

    resultsDiv.innerHTML = `
        <h3>Found ${data.total} result${data.total !== 1 ? 's' : ''} for "${query}"</h3>
        ${data.results.map(result => {
            // Find image metadata
            const img = allImages.find(i => i.id === result.image_id) || {};
            return `
                <div class="search-result" onclick="showImageModal('${result.image_id}')">
                    <img src="/data/${img.thumbnail_path || img.file_path || 'placeholder.jpg'}" alt="${img.title || 'Image'}"
                         onerror="this.src='data:image/svg+xml,<svg xmlns=%22http://www.w3.org/2000/svg%22 width=%22150%22 height=%22150%22><rect fill=%22%23ddd%22 width=%22150%22 height=%22150%22/></svg>'">
                    <div class="search-result-body">
                        <div class="search-result-score">Score: ${(result.relevance_score * 100).toFixed(0)}%</div>
                        <div class="search-result-title">${img.title || result.image_id.substring(0, 8)}</div>
                        <div class="search-result-category">${img.category || 'uncategorized'}</div>
                        ${img.description ? `<p>${img.description}</p>` : ''}
                        <div class="search-result-reason"><strong>Match reason:</strong> ${result.reason}</div>
                    </div>
                </div>
            `;
        }).join('')}
    `;
}

function searchExample(query) {
    document.getElementById('searchInput').value = query;
    performSearch();
}

// Search on Enter key
document.getElementById('searchInput').addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
        performSearch();
    }
});

// Load warehouse on page load
window.addEventListener('load', () => {
    loadWarehouse();
});
