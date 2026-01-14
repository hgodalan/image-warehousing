// Three.js ES module imports (must be at top of file)
import * as THREE from 'three';
import { OrbitControls } from 'three/addons/controls/OrbitControls.js';
import { STLLoader } from 'three/addons/loaders/STLLoader.js';
import { OBJLoader } from 'three/addons/loaders/OBJLoader.js';
import { FBXLoader } from 'three/addons/loaders/FBXLoader.js';

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

    grid.innerHTML = images.map(img => {
        // For 3D objects, use the front view as thumbnail
        let thumbnailPath = img.thumbnail_path || img.file_path;
        if (img.type === '3D' && img.views && img.views.front) {
            thumbnailPath = img.views.front;
        }

        return `
            <div class="image-card" onclick="showImageModal('${img.id}')">
                ${img.type === '3D' ? '<div class="badge-3d">3D</div>' : ''}
                <img src="/data/${thumbnailPath}" alt="${img.title}"
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
        `;
    }).join('');
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

        // Check if it's a 3D object
        if (img.type === '3D' && img.views) {
            const viewsHTML = Object.entries(img.views).map(([viewName, viewPath]) => `
                <div class="view-item">
                    <img src="/data/${viewPath}" alt="${viewName} view">
                    <div class="view-label">${viewName}</div>
                </div>
            `).join('');

            // Check file format and choose appropriate viewer
            const modelExt = img.model_filename ? img.model_filename.split('.').pop().toLowerCase() : '';
            const supportsModelViewer = ['glb', 'gltf'].includes(modelExt);
            const supportsThreeJs = ['stl', 'obj', 'fbx'].includes(modelExt);

            modalBody.innerHTML = `
                <h2>${img.title || 'Untitled'} <span class="badge-3d">3D</span></h2>
                <p><strong>Artist:</strong> ${img.artist || 'Unknown'}</p>
                <p><strong>Category:</strong> ${img.category || 'uncategorized'}</p>
                <p><strong>Uploaded:</strong> ${img.uploaded_at || 'N/A'}</p>

                ${img.model_file_path && supportsModelViewer ? `
                    <h3>Interactive 3D Preview:</h3>
                    <model-viewer
                        src="/data/${img.model_file_path}"
                        alt="${img.title || 'Untitled'}"
                        camera-controls
                        auto-rotate
                        style="width: 100%; height: 500px; background-color: #f5f5f5; border-radius: 8px; margin: 20px 0;"
                        loading="eager">
                        <div slot="progress-bar">
                            <div class="loading">Loading 3D model...</div>
                        </div>
                    </model-viewer>
                ` : img.model_file_path && supportsThreeJs ? `
                    <h3>Interactive 3D Preview:</h3>
                    <div id="threejs-viewer" style="width: 100%; height: 500px; background-color: #f5f5f5; border-radius: 8px; margin: 20px 0; position: relative;">
                        <div class="loading">Loading 3D model...</div>
                    </div>
                ` : ''}

                ${img.model_file_path ? `
                    <div class="download-section">
                        <a href="/data/${img.model_file_path}" download="${img.model_filename || 'model'}" class="btn download-btn">
                            📥 Download 3D Model (${img.model_filename || 'model'})
                        </a>
                        ${!supportsModelViewer && !supportsThreeJs ? '<p style="color: #999; margin-top: 10px; font-size: 0.9em;">Note: Interactive preview only available for .glb, .gltf, .stl, .obj, and .fbx files</p>' : ''}
                    </div>
                ` : ''}

                <h3>Surface Views:</h3>
                <div class="views-grid">
                    ${viewsHTML}
                </div>

                ${img.ai_analysis?.description || img.description ? `
                    <p><strong>Description:</strong> ${img.ai_analysis?.description || img.description}</p>
                ` : ''}
                ${img.manual_tags && img.manual_tags.length > 0 ? `
                    <p><strong>Tags:</strong> ${img.manual_tags.join(', ')}</p>
                ` : ''}
            `;

            // Initialize Three.js viewer if needed
            if (img.model_file_path && supportsThreeJs) {
                // Wait for DOM to update, then initialize viewer
                setTimeout(() => {
                    initThreeJsViewer('threejs-viewer', `/data/${img.model_file_path}`, modelExt);
                }, 100);
            }
        } else {
            // 2D image display
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
        }
    } catch (err) {
        modalBody.innerHTML = `<p>Error loading image: ${err.message}</p>`;
    }
}

function closeModal() {
    // Clean up Three.js viewer if it exists
    if (threeJsViewer) {
        threeJsViewer.renderer.dispose();
        threeJsViewer = null;
    }
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

// ============ 3D UPLOAD TAB ============

// State for 3D upload
let surfaceFiles = {};
let modelFile = null;

// Initialize 3D upload on page load
function init3DUpload() {
    const surfaceUploadsDiv = document.getElementById('surfaceUploads');
    const modeRadios = document.querySelectorAll('input[name="surfaceMode"]');
    const modelUploadBox = document.getElementById('modelUploadBox');
    const modelFileInput = document.getElementById('modelFileInput');
    const modelFileNameDiv = document.getElementById('modelFileName');

    // Handle mode change
    modeRadios.forEach(radio => {
        radio.addEventListener('change', (e) => {
            const mode = e.target.value;
            renderSurfaceSlots(mode);
        });
    });

    // Model file upload handlers
    modelUploadBox.addEventListener('click', () => {
        modelFileInput.click();
    });

    modelFileInput.addEventListener('change', (e) => {
        const file = e.target.files[0];
        if (file) {
            modelFile = file;
            modelUploadBox.classList.add('has-file');
            modelFileNameDiv.textContent = `📦 ${file.name} (${(file.size / 1024 / 1024).toFixed(2)} MB)`;
            modelFileNameDiv.classList.add('visible');
        }
    });

    // Initial render (4-surface mode)
    renderSurfaceSlots('4');

    // Upload button handler
    document.getElementById('upload3dBtn').addEventListener('click', upload3DObject);
}

function renderSurfaceSlots(mode) {
    const surfaceUploadsDiv = document.getElementById('surfaceUploads');
    const surfaces = mode === '4'
        ? ['front', 'back', 'left', 'right']
        : ['front', 'back', 'left', 'right', 'top', 'bottom'];

    // Clear existing files that don't exist in new mode
    const newSurfaceFiles = {};
    surfaces.forEach(surface => {
        if (surfaceFiles[surface]) {
            newSurfaceFiles[surface] = surfaceFiles[surface];
        }
    });
    surfaceFiles = newSurfaceFiles;

    // Render slots
    surfaceUploadsDiv.innerHTML = surfaces.map(surface => `
        <div class="surface-slot ${surfaceFiles[surface] ? 'has-file' : ''}" data-surface="${surface}">
            <div class="surface-slot-label">${surface.charAt(0).toUpperCase() + surface.slice(1)}</div>
            ${surfaceFiles[surface] ? `
                <img src="${URL.createObjectURL(surfaceFiles[surface])}" class="surface-slot-preview" alt="${surface}">
                <div class="surface-slot-filename">${surfaceFiles[surface].name}</div>
            ` : `
                <div class="surface-slot-placeholder">📷</div>
                <div class="surface-slot-filename">Click to select image</div>
            `}
            <input type="file" accept="image/*" data-surface="${surface}">
        </div>
    `).join('');

    // Add click handlers
    surfaceUploadsDiv.querySelectorAll('.surface-slot').forEach(slot => {
        const surface = slot.dataset.surface;
        const fileInput = slot.querySelector('input[type="file"]');

        slot.addEventListener('click', () => {
            fileInput.click();
        });

        fileInput.addEventListener('change', (e) => {
            const file = e.target.files[0];
            if (file) {
                surfaceFiles[surface] = file;
                renderSurfaceSlots(document.querySelector('input[name="surfaceMode"]:checked').value);
            }
        });
    });
}

async function upload3DObject() {
    const title = document.getElementById('upload3dTitle').value.trim();
    const artist = document.getElementById('upload3dArtist').value.trim();
    const tagsInput = document.getElementById('upload3dTags').value.trim();
    const mode = document.querySelector('input[name="surfaceMode"]:checked').value;
    const statusDiv = document.getElementById('upload3dStatus');

    // Validation
    if (!title || !artist) {
        statusDiv.innerHTML = '<div class="error">Title and Artist are required</div>';
        return;
    }

    if (!modelFile) {
        statusDiv.innerHTML = '<div class="error">3D model file is required</div>';
        return;
    }

    const requiredSurfaces = mode === '4'
        ? ['front', 'back', 'left', 'right']
        : ['front', 'back', 'left', 'right', 'top', 'bottom'];

    const missingSurfaces = requiredSurfaces.filter(s => !surfaceFiles[s]);
    if (missingSurfaces.length > 0) {
        statusDiv.innerHTML = `<div class="error">Missing surfaces: ${missingSurfaces.join(', ')}</div>`;
        return;
    }

    // Build FormData
    const formData = new FormData();
    formData.append('title', title);
    formData.append('artist', artist);
    formData.append('mode', mode);
    formData.append('model', modelFile); // Add the 3D model file

    if (tagsInput) {
        const tags = tagsInput.split(',').map(t => t.trim()).filter(t => t);
        formData.append('tags', JSON.stringify(tags));
    }

    // Add all surface files
    requiredSurfaces.forEach(surface => {
        formData.append(surface, surfaceFiles[surface]);
    });

    // Upload
    try {
        statusDiv.innerHTML = '<div class="info">Uploading 3D object...</div>';
        document.getElementById('upload3dBtn').disabled = true;

        const response = await fetch(`${API_BASE}/images/upload-3d`, {
            method: 'POST',
            body: formData
        });

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(errorText || `HTTP ${response.status}`);
        }

        const result = await response.json();

        statusDiv.innerHTML = `
            <div class="success">
                ✅ 3D object uploaded successfully!<br>
                <small>ID: ${result.id} | Views: ${result.views} | Status: ${result.status}</small>
            </div>
        `;

        // Clear form
        document.getElementById('upload3dTitle').value = '';
        document.getElementById('upload3dArtist').value = '';
        document.getElementById('upload3dTags').value = '';
        surfaceFiles = {};
        modelFile = null;
        document.getElementById('modelUploadBox').classList.remove('has-file');
        document.getElementById('modelFileName').classList.remove('visible');
        document.getElementById('modelFileInput').value = '';
        renderSurfaceSlots(mode);

    } catch (error) {
        statusDiv.innerHTML = `<div class="error">Upload failed: ${error.message}</div>`;
    } finally {
        document.getElementById('upload3dBtn').disabled = false;
    }
}

// ============ THREE.JS 3D VIEWER ============

let threeJsViewer = null;

window.initThreeJsViewer = function(containerId, modelUrl, fileExtension) {
    // Clean up previous viewer if exists
    if (threeJsViewer) {
        threeJsViewer.renderer.dispose();
        threeJsViewer = null;
    }

    const container = document.getElementById(containerId);
    if (!container) return;

    // Scene setup
    const scene = new THREE.Scene();
    scene.background = new THREE.Color(0xf5f5f5);

    // Camera
    const camera = new THREE.PerspectiveCamera(
        45,
        container.clientWidth / container.clientHeight,
        0.1,
        10000
    );
    camera.position.set(0, 50, 100);

    // Renderer
    const renderer = new THREE.WebGLRenderer({ antialias: true });
    renderer.setSize(container.clientWidth, container.clientHeight);
    renderer.setPixelRatio(window.devicePixelRatio);
    container.appendChild(renderer.domElement);

    // Lights
    const ambientLight = new THREE.AmbientLight(0xffffff, 0.6);
    scene.add(ambientLight);

    const directionalLight1 = new THREE.DirectionalLight(0xffffff, 0.8);
    directionalLight1.position.set(1, 1, 1);
    scene.add(directionalLight1);

    const directionalLight2 = new THREE.DirectionalLight(0xffffff, 0.5);
    directionalLight2.position.set(-1, -1, -1);
    scene.add(directionalLight2);

    // Controls
    const controls = new OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true;
    controls.dampingFactor = 0.05;
    controls.autoRotate = true;
    controls.autoRotateSpeed = 2.0;

    // Load model based on file type
    const loadModel = () => {
        const ext = fileExtension.toLowerCase();

        if (ext === 'stl') {
            const loader = new STLLoader();
            loader.load(
                modelUrl,
                (geometry) => {
                    const material = new THREE.MeshPhongMaterial({
                        color: 0x667eea,
                        specular: 0x111111,
                        shininess: 200
                    });
                    const mesh = new THREE.Mesh(geometry, material);

                    // Center and scale the model
                    geometry.computeBoundingBox();
                    const center = new THREE.Vector3();
                    geometry.boundingBox.getCenter(center);
                    mesh.position.sub(center);

                    const box = new THREE.Box3().setFromObject(mesh);
                    const size = box.getSize(new THREE.Vector3()).length();
                    const scale = 50 / size;
                    mesh.scale.setScalar(scale);

                    scene.add(mesh);
                },
                undefined,
                (error) => {
                    console.error('Error loading STL:', error);
                    container.innerHTML = '<div class="error">Failed to load 3D model</div>';
                }
            );
        } else if (ext === 'obj') {
            const loader = new OBJLoader();
            loader.load(
                modelUrl,
                (object) => {
                    // Apply material to all meshes
                    object.traverse((child) => {
                        if (child.isMesh) {
                            child.material = new THREE.MeshPhongMaterial({
                                color: 0x667eea,
                                specular: 0x111111,
                                shininess: 200
                            });
                        }
                    });

                    // Center and scale
                    const box = new THREE.Box3().setFromObject(object);
                    const center = box.getCenter(new THREE.Vector3());
                    object.position.sub(center);

                    const size = box.getSize(new THREE.Vector3()).length();
                    const scale = 50 / size;
                    object.scale.setScalar(scale);

                    scene.add(object);
                },
                undefined,
                (error) => {
                    console.error('Error loading OBJ:', error);
                    container.innerHTML = '<div class="error">Failed to load 3D model</div>';
                }
            );
        } else if (ext === 'fbx') {
            const loader = new FBXLoader();
            loader.load(
                modelUrl,
                (object) => {
                    // Center and scale
                    const box = new THREE.Box3().setFromObject(object);
                    const center = box.getCenter(new THREE.Vector3());
                    object.position.sub(center);

                    const size = box.getSize(new THREE.Vector3()).length();
                    const scale = 50 / size;
                    object.scale.setScalar(scale);

                    scene.add(object);
                },
                undefined,
                (error) => {
                    console.error('Error loading FBX:', error);
                    container.innerHTML = '<div class="error">Failed to load 3D model</div>';
                }
            );
        }
    };

    loadModel();

    // Animation loop
    function animate() {
        requestAnimationFrame(animate);
        controls.update();
        renderer.render(scene, camera);
    }
    animate();

    // Handle window resize
    window.addEventListener('resize', () => {
        if (container.clientWidth > 0) {
            camera.aspect = container.clientWidth / container.clientHeight;
            camera.updateProjectionMatrix();
            renderer.setSize(container.clientWidth, container.clientHeight);
        }
    });

    threeJsViewer = { scene, camera, renderer, controls };
}

// Expose functions to global window for inline event handlers
window.loadWarehouse = loadWarehouse;
window.performSearch = performSearch;
window.searchExample = searchExample;
window.closeModal = closeModal;
window.showImageModal = showImageModal;

// Load warehouse on page load
window.addEventListener('load', () => {
    loadWarehouse();
    init3DUpload();
});
