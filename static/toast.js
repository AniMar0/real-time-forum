let toastContainer

function initToastContainer() {
  if (!toastContainer) {
    toastContainer = document.createElement("div")
    toastContainer.id = "toast-container"
    document.body.appendChild(toastContainer)
  }
  return toastContainer
}

export function showToast(message, type = "success", duration = 4000) {
  const container = initToastContainer()

  const toast = document.createElement("div")
  toast.className = `toast ${type}`

  const icon = type === "success" ? "✓" : "✕"
  const title = type === "success" ? "Success" : "Error"

  toast.innerHTML = `
    <div class="toast-icon">${icon}</div>
    <div class="toast-content">
      <p class="toast-title">${title}</p>
      <p class="toast-message">${message}</p>
    </div>
    <button class="toast-close" aria-label="Close">&times;</button>
    <div class="toast-progress" style="animation-duration: ${duration}ms;"></div>
  `

  container.appendChild(toast)

  const closeBtn = toast.querySelector(".toast-close")
  closeBtn.addEventListener("click", () => removeToast(toast))

  const timeout = setTimeout(() => {
    removeToast(toast)
  }, duration)

  toast.addEventListener("mouseenter", () => {
    clearTimeout(timeout)
    const progress = toast.querySelector(".toast-progress")
    if (progress) {
      progress.style.animationPlayState = "paused"
    }
  })

  toast.addEventListener("mouseleave", () => {
    const progress = toast.querySelector(".toast-progress")
    if (progress) {
      const remainingTime = (Number.parseFloat(getComputedStyle(progress).width) / toast.offsetWidth) * duration
      progress.style.animationPlayState = "running"
      setTimeout(() => removeToast(toast), remainingTime)
    }
  })

  return toast
}

function removeToast(toast) {
  if (!toast || !toast.parentElement) return

  toast.classList.add("removing")
  setTimeout(() => {
    if (toast.parentElement) {
      toast.parentElement.removeChild(toast)
    }
  }, 300)
}

export function successToast(message, duration) {
  return showToast(message, "success", duration)
}

export function errorToast(message, duration) {
  return showToast(message, "error", duration)
}
