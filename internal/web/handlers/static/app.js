// E-Ink 6-Color Image Processor - Frontend Application

class EInkProcessor {
    constructor() {
        this.currentImage = null;
        this.currentImageData = null;
        this.crop = { x: 0.1, y: 0.1, width: 0.8, height: 0.8 }; // Default crop (relative 0-1)
        this.targetAspect = 800 / 480; // Default landscape aspect
        this.isLandscape = true; // Track orientation
        this.init();
    }

    init() {
        // Cache DOM elements
        this.uploadArea = document.getElementById('uploadArea');
        this.fileInput = document.getElementById('fileInput');
        this.previewSection = document.getElementById('previewSection');
        this.controlsSection = document.getElementById('controlsSection');
        this.stepsSection = document.getElementById('stepsSection');
        this.remoteSection = document.getElementById('remoteSection');
        this.originalImage = document.getElementById('originalImage');
        this.previewImage = document.getElementById('previewImage');
        this.originalInfo = document.getElementById('originalInfo');
        this.previewInfo = document.getElementById('previewInfo');
        this.stepsContainer = document.getElementById('stepsContainer');
        this.loadingOverlay = document.getElementById('loadingOverlay');

        // Crop elements
        this.cropContainer = document.getElementById('cropContainer');
        this.cropOverlay = document.getElementById('cropOverlay');
        this.cropBorder = document.getElementById('cropBorder');
        this.cropSizeInfo = document.getElementById('cropSizeInfo');

        // Orientation toggle
        this.orientationBtn = document.getElementById('orientationBtn');
        this.orientationLabel = document.getElementById('orientationLabel');

        // Sliders and inputs
        this.resolutionPreset = document.getElementById('resolutionPreset');
        this.outputFormat = document.getElementById('outputFormat');
        this.widthInput = document.getElementById('widthInput');
        this.heightInput = document.getElementById('heightInput');
        this.brightnessSlider = document.getElementById('brightnessSlider');
        this.contrastSlider = document.getElementById('contrastSlider');
        this.saturationSlider = document.getElementById('saturationSlider');
        this.ditherToggle = document.getElementById('ditherToggle');
        this.enhancerSelect = document.getElementById('enhancerSelect');
        this.remoteUrlInput = document.getElementById('remoteUrlInput');

        // Buttons
        this.updateBtn = document.getElementById('updateBtn');
        this.uploadBtn = document.getElementById('uploadBtn');
        this.uploadProtocolSelect = document.getElementById('uploadProtocolSelect');

        // Setup event listeners
        this.setupEventListeners();
        this.setupCropListeners();

        // Load enhancers
        this.loadEnhancers();
    }

    async loadEnhancers() {
        try {
            const response = await fetch('/api/enhancers');
            const data = await response.json();

            if (data.success && data.enhancers) {
                // Keep the first "Basic" option, add others
                data.enhancers.forEach(enhancer => {
                    // Skip basic as it's already the default option
                    if (enhancer.name === 'basic') return;

                    const option = document.createElement('option');
                    option.value = enhancer.name;
                    option.textContent = `${enhancer.displayName} - ${enhancer.description}`;
                    this.enhancerSelect.appendChild(option);
                });
            }
        } catch (error) {
            console.error('Failed to load enhancers:', error);
        }
    }

    setupEventListeners() {
        // Upload area events
        this.uploadArea.addEventListener('click', () => this.fileInput.click());
        this.fileInput.addEventListener('change', (e) => this.handleFileSelect(e));

        // Drag and drop
        this.uploadArea.addEventListener('dragover', (e) => {
            e.preventDefault();
            this.uploadArea.classList.add('dragover');
        });

        this.uploadArea.addEventListener('dragleave', () => {
            this.uploadArea.classList.remove('dragover');
        });

        this.uploadArea.addEventListener('drop', (e) => {
            e.preventDefault();
            this.uploadArea.classList.remove('dragover');
            const file = e.dataTransfer.files[0];
            if (file && file.type.startsWith('image/')) {
                this.loadImage(file);
            }
        });

        // Slider value updates
        this.brightnessSlider.addEventListener('input', (e) => {
            document.getElementById('brightnessValue').textContent = e.target.value;
        });

        this.contrastSlider.addEventListener('input', (e) => {
            document.getElementById('contrastValue').textContent = parseFloat(e.target.value).toFixed(2);
        });

        this.saturationSlider.addEventListener('input', (e) => {
            document.getElementById('saturationValue').textContent = parseFloat(e.target.value).toFixed(2);
        });

        // Resolution preset
        this.resolutionPreset.addEventListener('change', (e) => this.handleResolutionPreset(e));

        // Orientation toggle
        this.orientationBtn.addEventListener('click', () => this.toggleOrientation());

        // Button events
        this.updateBtn.addEventListener('click', () => this.updatePreview());
        this.uploadBtn.addEventListener('click', () => this.uploadToDisplay());

        // Enter key on remote URL input
        this.remoteUrlInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.updatePreview();
            }
        });
    }

    toggleOrientation() {
        this.isLandscape = !this.isLandscape;

        if (this.isLandscape) {
            this.targetAspect = 800 / 480;
            this.widthInput.value = 800;
            this.heightInput.value = 480;
            this.orientationLabel.textContent = '800×480';
        } else {
            this.targetAspect = 480 / 800;
            this.widthInput.value = 480;
            this.heightInput.value = 800;
            this.orientationLabel.textContent = '480×800';
        }

        // Adjust crop to new aspect ratio
        this.adjustCropForAspect();
    }

    setupCropListeners() {
        if (!this.cropOverlay) return;

        let isDragging = false;
        let isResizing = false;
        let activeHandle = null;
        let startX, startY;
        let startCrop = { ...this.crop };

        // Deactivate on outside click
        document.addEventListener('click', (e) => {
            if (!this.cropOverlay.contains(e.target)) {
                this.cropOverlay.classList.remove('active');
            }
        });

        // Mouse down on overlay or border - start dragging
        this.cropBorder.addEventListener('mousedown', (e) => {
            if (e.target.classList.contains('crop-handle')) {
                // Resizing with handle
                isResizing = true;
                activeHandle = e.target.dataset.handle;
            } else {
                // Dragging the whole crop area
                isDragging = true;
                this.cropOverlay.classList.add('active');
            }

            startX = e.clientX;
            startY = e.clientY;
            startCrop = { ...this.crop };
            e.preventDefault();
            e.stopPropagation();
        });

        document.addEventListener('mousemove', (e) => {
            if (!isDragging && !isResizing) return;

            if (isDragging) {
                // Move the entire crop area using image-relative coordinates
                const imgRect = this.originalImage.getBoundingClientRect();
                const mouseX = (e.clientX - imgRect.left) / imgRect.width;
                const mouseY = (e.clientY - imgRect.top) / imgRect.height;
                const startMouseX = (startX - imgRect.left) / imgRect.width;
                const startMouseY = (startY - imgRect.top) / imgRect.height;

                const deltaX = mouseX - startMouseX;
                const deltaY = mouseY - startMouseY;

                this.crop.x = Math.max(0, Math.min(1 - startCrop.width, startCrop.x + deltaX));
                this.crop.y = Math.max(0, Math.min(1 - startCrop.height, startCrop.y + deltaY));
                this.updateCropVisual();
            } else if (isResizing) {
                // Resize from handle - pass the mouse event directly
                this.handleResize(activeHandle, e, startCrop);
            }
        });

        document.addEventListener('mouseup', () => {
            isDragging = false;
            isResizing = false;
            activeHandle = null;
        });
    }

    handleResize(handle, mouseEvent, startCrop) {
        if (!this.currentImage) return;

        // Get the image element's bounding rect for mouse position calculation
        const imgRect = this.originalImage.getBoundingClientRect();

        // Calculate mouse position relative to image, normalized to [0, 1]
        const mouseX = Math.max(0, Math.min(1, (mouseEvent.clientX - imgRect.left) / imgRect.width));
        const mouseY = Math.max(0, Math.min(1, (mouseEvent.clientY - imgRect.top) / imgRect.height));

        // Get target aspect ratio and original image dimensions
        const targetW = parseInt(this.widthInput.value) || 800;
        const targetH = parseInt(this.heightInput.value) || 480;
        const targetAspect = targetW / targetH; // This is the aspect ratio we want to maintain

        // Get original image actual pixel dimensions
        const imgPixelW = this.currentImage.width;
        const imgPixelH = this.currentImage.height;

        const { x, y, width, height } = startCrop;
        let newX, newY, newWidth, newHeight;

        // Convert normalized coordinates to pixel coordinates for calculation
        const toPixel = (normX, normY, normW, normH) => ({
            x: normX * imgPixelW,
            y: normY * imgPixelH,
            width: normW * imgPixelW,
            height: normH * imgPixelH
        });

        const toNorm = (pixelX, pixelY, pixelW, pixelH) => ({
            x: pixelX / imgPixelW,
            y: pixelY / imgPixelH,
            width: pixelW / imgPixelW,
            height: pixelH / imgPixelH
        });

        // Convert mouse position to pixels
        const mousePixelX = mouseX * imgPixelW;
        const mousePixelY = mouseY * imgPixelH;

        // Convert current crop to pixels
        const cropPixel = toPixel(x, y, width, height);

        // Each handle has a fixed corner and a moving corner
        switch (handle) {
            case 'se': {
                // Fixed: top-left, Moving: bottom-right
                const fixedPixelX = cropPixel.x;
                const fixedPixelY = cropPixel.y;

                // Calculate proposed dimensions in pixels
                let proposedPixelW = mousePixelX - fixedPixelX;
                let proposedPixelH = mousePixelY - fixedPixelY;

                // Calculate actual dimensions maintaining target aspect ratio
                // We'll use whichever direction the user is dragging more
                const heightFromWidth = proposedPixelW / targetAspect;
                const widthFromHeight = proposedPixelH * targetAspect;

                let finalPixelW, finalPixelH;
                if (Math.abs(proposedPixelW - cropPixel.width) > Math.abs(proposedPixelH - cropPixel.height)) {
                    // User is dragging more horizontally
                    finalPixelW = Math.max(50, proposedPixelW);
                    finalPixelH = finalPixelW / targetAspect;
                } else {
                    // User is dragging more vertically
                    finalPixelH = Math.max(50, proposedPixelH);
                    finalPixelW = finalPixelH * targetAspect;
                }

                // Clamp to image boundaries
                if (fixedPixelX + finalPixelW > imgPixelW) {
                    finalPixelW = imgPixelW - fixedPixelX;
                    finalPixelH = finalPixelW / targetAspect;
                }
                if (fixedPixelY + finalPixelH > imgPixelH) {
                    finalPixelH = imgPixelH - fixedPixelY;
                    finalPixelW = finalPixelH * targetAspect;
                }

                const result = toNorm(fixedPixelX, fixedPixelY, finalPixelW, finalPixelH);
                newX = result.x;
                newY = result.y;
                newWidth = result.width;
                newHeight = result.height;
                break;
            }
            case 'sw': {
                // Fixed: top-right, Moving: bottom-left
                const fixedPixelX = cropPixel.x + cropPixel.width;
                const fixedPixelY = cropPixel.y;

                let proposedPixelW = fixedPixelX - mousePixelX;
                let proposedPixelH = mousePixelY - fixedPixelY;

                let finalPixelW, finalPixelH;
                if (Math.abs(proposedPixelW - cropPixel.width) > Math.abs(proposedPixelH - cropPixel.height)) {
                    finalPixelW = Math.max(50, proposedPixelW);
                    finalPixelH = finalPixelW / targetAspect;
                } else {
                    finalPixelH = Math.max(50, proposedPixelH);
                    finalPixelW = finalPixelH * targetAspect;
                }

                let newPixelX = fixedPixelX - finalPixelW;
                if (newPixelX < 0) {
                    newPixelX = 0;
                    finalPixelW = fixedPixelX;
                    finalPixelH = finalPixelW / targetAspect;
                }
                if (fixedPixelY + finalPixelH > imgPixelH) {
                    finalPixelH = imgPixelH - fixedPixelY;
                    finalPixelW = finalPixelH * targetAspect;
                    newPixelX = fixedPixelX - finalPixelW;
                }

                const result = toNorm(newPixelX, fixedPixelY, finalPixelW, finalPixelH);
                newX = result.x;
                newY = result.y;
                newWidth = result.width;
                newHeight = result.height;
                break;
            }
            case 'ne': {
                // Fixed: bottom-left, Moving: top-right
                const fixedPixelX = cropPixel.x;
                const fixedPixelY = cropPixel.y + cropPixel.height;

                let proposedPixelW = mousePixelX - fixedPixelX;
                let proposedPixelH = fixedPixelY - mousePixelY;

                let finalPixelW, finalPixelH;
                if (Math.abs(proposedPixelW - cropPixel.width) > Math.abs(proposedPixelH - cropPixel.height)) {
                    finalPixelW = Math.max(50, proposedPixelW);
                    finalPixelH = finalPixelW / targetAspect;
                } else {
                    finalPixelH = Math.max(50, proposedPixelH);
                    finalPixelW = finalPixelH * targetAspect;
                }

                if (fixedPixelX + finalPixelW > imgPixelW) {
                    finalPixelW = imgPixelW - fixedPixelX;
                    finalPixelH = finalPixelW / targetAspect;
                }

                let newPixelY = fixedPixelY - finalPixelH;
                if (newPixelY < 0) {
                    newPixelY = 0;
                    finalPixelH = fixedPixelY;
                    finalPixelW = finalPixelH * targetAspect;
                }

                const result = toNorm(fixedPixelX, newPixelY, finalPixelW, finalPixelH);
                newX = result.x;
                newY = result.y;
                newWidth = result.width;
                newHeight = result.height;
                break;
            }
            case 'nw': {
                // Fixed: bottom-right, Moving: top-left
                const fixedPixelX = cropPixel.x + cropPixel.width;
                const fixedPixelY = cropPixel.y + cropPixel.height;

                let proposedPixelW = fixedPixelX - mousePixelX;
                let proposedPixelH = fixedPixelY - mousePixelY;

                let finalPixelW, finalPixelH;
                if (Math.abs(proposedPixelW - cropPixel.width) > Math.abs(proposedPixelH - cropPixel.height)) {
                    finalPixelW = Math.max(50, proposedPixelW);
                    finalPixelH = finalPixelW / targetAspect;
                } else {
                    finalPixelH = Math.max(50, proposedPixelH);
                    finalPixelW = finalPixelH * targetAspect;
                }

                let newPixelX = fixedPixelX - finalPixelW;
                let newPixelY = fixedPixelY - finalPixelH;

                if (newPixelX < 0) {
                    newPixelX = 0;
                    finalPixelW = fixedPixelX;
                    finalPixelH = finalPixelW / targetAspect;
                    newPixelY = fixedPixelY - finalPixelH;
                }
                if (newPixelY < 0) {
                    newPixelY = 0;
                    finalPixelH = fixedPixelY;
                    finalPixelW = finalPixelH * targetAspect;
                    newPixelX = fixedPixelX - finalPixelW;
                }

                const result = toNorm(newPixelX, newPixelY, finalPixelW, finalPixelH);
                newX = result.x;
                newY = result.y;
                newWidth = result.width;
                newHeight = result.height;
                break;
            }
            default:
                return;
        }

        // Enforce minimum size in normalized coordinates
        if (newWidth < 0.05 || newHeight < 0.05) {
            return;
        }

        // Update crop
        this.crop.x = newX;
        this.crop.y = newY;
        this.crop.width = newWidth;
        this.crop.height = newHeight;

        this.updateCropVisual();
    }

    updateCropVisual() {
        if (!this.cropBorder) return;

        this.cropBorder.style.left = (this.crop.x * 100) + '%';
        this.cropBorder.style.top = (this.crop.y * 100) + '%';
        this.cropBorder.style.width = (this.crop.width * 100) + '%';
        this.cropBorder.style.height = (this.crop.height * 100) + '%';
        this.cropBorder.style.transform = 'none';

        // Update pixel size display
        this.updateCropSizeInfo();
    }

    updateCropSizeInfo() {
        if (!this.cropSizeInfo || !this.currentImage) return;

        // Calculate actual pixels based on original image dimensions
        const pixelWidth = Math.round(this.currentImage.width * this.crop.width);
        const pixelHeight = Math.round(this.currentImage.height * this.crop.height);

        // Show crop percentage as well for debugging
        const widthPct = Math.round(this.crop.width * 100);
        const heightPct = Math.round(this.crop.height * 100);

        this.cropSizeInfo.textContent = `${pixelWidth} × ${pixelHeight} (${widthPct}% × ${heightPct}%)`;
    }

    handleResolutionPreset(e) {
        const preset = e.target.value;
        switch (preset) {
            case '800x480':
                this.widthInput.value = 800;
                this.heightInput.value = 480;
                this.targetAspect = 800 / 480;
                break;
            case '480x800':
                this.widthInput.value = 480;
                this.heightInput.value = 800;
                this.targetAspect = 480 / 800;
                break;
            case 'custom':
                // Keep current values
                break;
        }
        this.adjustCropForAspect();
    }

    adjustCropForAspect() {
        if (!this.currentImage) return;

        // Calculate crop box as maximum multiple of target size that fits in image
        const imgW = this.currentImage.width;
        const imgH = this.currentImage.height;
        const targetW = parseInt(this.widthInput.value) || 800;
        const targetH = parseInt(this.heightInput.value) || 480;

        // Calculate the maximum scale factor that fits within the image
        const scaleX = imgW / targetW;  // How many target widths fit in image width
        const scaleY = imgH / targetH;  // How many target heights fit in image height
        const scale = Math.min(scaleX, scaleY);  // Use the smaller to fit both dimensions

        // Calculate crop size (in pixels and as percentage)
        const cropPixelW = Math.round(targetW * scale);
        const cropPixelH = Math.round(targetH * scale);
        const newCrop = {
            width: cropPixelW / imgW,
            height: cropPixelH / imgH,
            x: 0,
            y: 0
        };

        // Center the crop
        newCrop.x = (1 - newCrop.width) / 2;
        newCrop.y = (1 - newCrop.height) / 2;

        this.crop = newCrop;
        this.updateCropVisual();
    }

    handleFileSelect(event) {
        const file = event.target.files[0];
        if (file) {
            this.loadImage(file);
        }
    }

    loadImage(file) {
        const reader = new FileReader();
        reader.onload = (e) => {
            this.currentImageData = e.target.result;
            this.originalImage.src = this.currentImageData;

            // Show sections
            this.previewSection.style.display = 'block';
            this.controlsSection.style.display = 'block';
            this.remoteSection.style.display = 'block';

            // Get image dimensions FIRST, then update preview
            const img = new Image();
            img.onload = () => {
                // Store image reference for crop size calculation
                this.currentImage = img;
                this.originalInfo.textContent = `${img.width}x${img.height}px`;

                // Set resolution based on image aspect
                const imageAspect = img.width / img.height;
                if (imageAspect > 1) {
                    // Landscape - use 800x480
                    this.isLandscape = true;
                    this.resolutionPreset.value = '800x480';
                    this.widthInput.value = 800;
                    this.heightInput.value = 480;
                    this.targetAspect = 800 / 480;
                } else {
                    // Portrait - use 480x800
                    this.isLandscape = false;
                    this.resolutionPreset.value = '480x800';
                    this.widthInput.value = 480;
                    this.heightInput.value = 800;
                    this.targetAspect = 480 / 800;
                }

                // Update orientation button label
                if (this.orientationLabel) {
                    this.orientationLabel.textContent = this.isLandscape ? '800×480' : '480×800';
                }

                // Adjust crop after image loads (with small delay for DOM)
                setTimeout(() => this.adjustCropForAspect(), 100);

                // NOW update preview after dimensions are set
                this.updatePreview();
            };
            img.src = this.currentImageData;
        };
        reader.readAsDataURL(file);
    }

    async updatePreview() {
        if (!this.currentImageData) {
            this.showToast('Please upload an image first', 'error');
            return;
        }

        this.showLoading(true);
        this.updateBtn.disabled = true;

        try {
            const config = {
                image: this.currentImageData.split(',')[1],
                width: parseInt(this.widthInput.value) || 0,
                height: parseInt(this.heightInput.value) || 0,
                cropX: this.crop.x,
                cropY: this.crop.y,
                cropWidth: this.crop.width,
                cropHeight: this.crop.height,
                outputFormat: this.outputFormat.value,
                brightness: parseFloat(this.brightnessSlider.value),
                contrast: parseFloat(this.contrastSlider.value),
                saturation: parseFloat(this.saturationSlider.value),
                dither: this.ditherToggle.checked,
                enhancerName: this.enhancerSelect.value
            };

            const response = await fetch('/api/preview', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(config)
            });

            const data = await response.json();

            if (data.success && data.result) {
                // Update preview image
                const mimeType = this.outputFormat.value === 'bmp' ? 'image/bmp' : 'image/png';
                this.previewImage.src = `data:${mimeType};base64,${data.result.final}`;

                // Get preview dimensions
                const img = new Image();
                img.onload = () => {
                    this.previewInfo.textContent = `${img.width}x${img.height}px`;
                };
                img.src = `data:${mimeType};base64,${data.result.final}`;

                // Show processing steps
                if (data.result.steps && data.result.steps.length > 0) {
                    this.displaySteps(data.result.steps);
                    this.stepsSection.style.display = 'block';
                }

                this.showToast('Preview updated', 'success');
            } else {
                throw new Error(data.error || 'Processing failed');
            }
        } catch (error) {
            console.error('Error updating preview:', error);
            this.showToast('Failed to update preview: ' + error.message, 'error');
        } finally {
            this.showLoading(false);
            this.updateBtn.disabled = false;
        }
    }

    displaySteps(steps) {
        this.stepsContainer.innerHTML = '';

        steps.forEach(step => {
            const stepDiv = document.createElement('div');
            stepDiv.className = 'step-item';

            const title = document.createElement('h4');
            title.textContent = step.name;

            const img = document.createElement('img');
            img.src = 'data:image/png;base64,' + step.image;
            img.alt = step.name;

            stepDiv.appendChild(title);
            stepDiv.appendChild(img);
            this.stepsContainer.appendChild(stepDiv);
        });
    }

    async uploadToDisplay() {
        if (!this.currentImageData) {
            this.showToast('Please upload an image first', 'error');
            return;
        }

        this.showLoading(true);
        this.uploadBtn.disabled = true;

        try {
            const config = {
                image: this.currentImageData.split(',')[1],
                width: parseInt(this.widthInput.value) || 0,
                height: parseInt(this.heightInput.value) || 0,
                cropX: this.crop.x,
                cropY: this.crop.y,
                cropWidth: this.crop.width,
                cropHeight: this.crop.height,
                outputFormat: this.outputFormat.value,
                brightness: parseFloat(this.brightnessSlider.value),
                contrast: parseFloat(this.contrastSlider.value),
                saturation: parseFloat(this.saturationSlider.value),
                dither: this.ditherToggle.checked,
                enhancerName: this.enhancerSelect.value,
                remoteUrl: this.remoteUrlInput.value || null,
                protocol: this.uploadProtocolSelect ? this.uploadProtocolSelect.value : 'sta'
            };

            const response = await fetch('/api/upload', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(config)
            });

            const data = await response.json();

            if (data.success) {
                this.showToast('Image uploaded to display!', 'success');
            } else {
                throw new Error(data.error || 'Upload failed');
            }
        } catch (error) {
            console.error('Error uploading:', error);
            this.showToast('Upload failed: ' + error.message, 'error');
        } finally {
            this.showLoading(false);
            this.uploadBtn.disabled = false;
        }
    }

    showLoading(show) {
        this.loadingOverlay.style.display = show ? 'flex' : 'none';
    }

    showToast(message, type = 'info') {
        const toast = document.getElementById('toast');
        toast.textContent = message;
        toast.className = 'toast ' + type;
        toast.classList.add('show');

        setTimeout(() => {
            toast.classList.remove('show');
        }, 3000);
    }
}

// Initialize the application when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    new EInkProcessor();
});
