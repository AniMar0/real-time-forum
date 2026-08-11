// Toast notification system

let toastContainer;

// Initialize toast container
function initToastContainer() {
    if (!toastContainer) {
        toastContainer = document.createElement('div');
        toastContainer.id = 'toast-container';
        document.body.appendChild(toastContainer);
    }
    return toastContainer;
}

// Show toast notification
export function showToast(message, type = 'success', duration = 4000) {
    const container = initToastContainer();

    // Create toast element
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;

    // Determine icon and title based on type
    const icon = type === 'success' ? '✓' : '✕';
    const title = type === 'success' ? 'Success' : 'Error';

    // Build toast DOM with textContent so server/user messages cannot inject HTML.
    const iconElement = document.createElement('div');
    iconElement.className = 'toast-icon';
    iconElement.textContent = icon;

    const content = document.createElement('div');
    content.className = 'toast-content';

    const titleElement = document.createElement('p');
    titleElement.className = 'toast-title';
    titleElement.textContent = title;

    const messageElement = document.createElement('p');
    messageElement.className = 'toast-message';
    messageElement.textContent = String(message);
    content.append(titleElement, messageElement);

    const closeButton = document.createElement('button');
    closeButton.className = 'toast-close';
    closeButton.type = 'button';
    closeButton.setAttribute('aria-label', 'Close');
    closeButton.textContent = '×';

    const progress = document.createElement('div');
    progress.className = 'toast-progress';
    progress.style.animationDuration = `${duration}ms`;
    toast.append(iconElement, content, closeButton, progress);

    // Add to container
    container.appendChild(toast);

    // Close button handler
    closeButton.addEventListener('click', () => removeToast(toast));

    // Auto-remove after duration
    const timeout = setTimeout(() => {
        removeToast(toast);
    }, duration);

    // Pause on hover
    toast.addEventListener('mouseenter', () => {
        clearTimeout(timeout);
        const progress = toast.querySelector('.toast-progress');
        if (progress) {
            progress.style.animationPlayState = 'paused';
        }
    });

    // Resume on mouse leave
    toast.addEventListener('mouseleave', () => {
        const progress = toast.querySelector('.toast-progress');
        if (progress) {
            const remainingTime = (parseFloat(getComputedStyle(progress).width) / toast.offsetWidth) * duration;
            progress.style.animationPlayState = 'running';
            setTimeout(() => removeToast(toast), remainingTime);
        }
    });

    return toast;
}

// Remove toast with animation
function removeToast(toast) {
    if (!toast || !toast.parentElement) return;

    toast.classList.add('removing');
    setTimeout(() => {
        if (toast.parentElement) {
            toast.parentElement.removeChild(toast);
        }
    }, 300);
}

// Convenience methods
export function successToast(message, duration) {
    return showToast(message, 'success', duration);
}

export function errorToast(message, duration) {
    return showToast(message, 'error', duration);
}
