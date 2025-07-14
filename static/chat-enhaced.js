const unreadCounts = new Map()
let socket = null
let selectedUser = null
let currentUser = null

export function startChatFeature(currentUsername) {
  currentUser = currentUsername

  socket = new WebSocket("ws://" + window.location.host + "/ws")

  socket.addEventListener("message", (event) => {
    const data = JSON.parse(event.data)

    if (data.type === "user_list") {
      setUserList(data.users)
    } else {
      if (data.from === selectedUser || data.to === selectedUser) {
        renderMessage(data)
      } else if (data.to === currentUser) {
        // Increase unread count
        const prev = unreadCounts.get(data.from) || 0
        unreadCounts.set(data.from, prev + 1)
        updateNotificationBadge(data.from)
      }
    }
  })

  // Add close button functionality
  addChatCloseButton()

  const sendBtn = document.getElementById("sendBtn")
  const input = document.getElementById("messageInput")

  if (sendBtn && input) {
    sendBtn.addEventListener("click", () => {
      const content = input.value.trim()
      if (!content || !selectedUser) return

      const message = {
        to: selectedUser,
        from: currentUser,
        content: content,
        timestamp: new Date().toISOString(),
      }

      socket.send(JSON.stringify(message))
      renderMessage(message)
      input.value = ""
    })

    // Add Enter key support
    input.addEventListener("keydown", (e) => {
      if (e.key === "Enter" && !e.shiftKey) {
        e.preventDefault()
        sendBtn.click()
      }
    })
  }
}

function addChatCloseButton() {
  const chatHeader = document.getElementById("chatHeader")
  if (chatHeader && !chatHeader.querySelector("#chatCloseBtn")) {
    const closeBtn = document.createElement("button")
    closeBtn.id = "chatCloseBtn"
    closeBtn.innerHTML = "×"
    closeBtn.title = "Close chat"

    closeBtn.addEventListener("click", () => {
      const chatWindow = document.getElementById("chatWindow")
      chatWindow.classList.add("hidden")
      selectedUser = null
    })

    chatHeader.appendChild(closeBtn)
  }
}

function renderMessage(msg) {
  const container = document.getElementById("chatMessages")
  const div = document.createElement("div")
  div.innerHTML = `
    <p><strong>${msg.from}</strong>: ${msg.content}<br/>
    <small>${new Date(msg.timestamp).toLocaleTimeString()}</small></p>
  `
  container.appendChild(div)

  // Smooth scroll to bottom
  container.scrollTo({
    top: container.scrollHeight,
    behavior: "smooth",
  })
}

function setUserList(users) {
  const list = document.getElementById("userList")
  list.innerHTML = "<h2>Online Users</h2>"

  users.forEach((username) => {
    if (username === currentUser) return

    const div = document.createElement("div")
    div.className = "user"
    div.style.display = "flex"
    div.style.justifyContent = "space-between"
    div.style.alignItems = "center"
    div.style.cursor = "pointer"

    const nameSpan = document.createElement("span")
    nameSpan.textContent = username

    const statusSpan = document.createElement("span")
    statusSpan.classList.add("status", "online")

    div.appendChild(nameSpan)
    div.appendChild(statusSpan)

    div.addEventListener("click", async () => {
      selectedUser = username
      document.getElementById("chatWithName").textContent = username
      document.getElementById("chatWindow").classList.remove("hidden")
      document.getElementById("chatMessages").innerHTML = ""

      unreadCounts.set(username, 0)
      const badge = div.querySelector(".notification-badge")
      if (badge) badge.remove()

      try {
        const res = await fetch(`/messages?from=${currentUser}&to=${selectedUser}`)
        if (!res.ok) throw new Error("Failed to load chat history")
        const messages = await res.json()
        messages.forEach(renderMessage)
      } catch (err) {
        console.error("Chat history error:", err)
      }
    })

    list.appendChild(div)
  })
}

function updateNotificationBadge(fromUser) {
  const userList = document.getElementById("userList")
  const users = userList.getElementsByClassName("user")

  for (const div of users) {
    const nameSpan = div.querySelector("span:first-child")
    if (nameSpan && nameSpan.textContent === fromUser) {
      let badge = div.querySelector(".notification-badge")

      // Create badge if it doesn't exist
      if (!badge) {
        badge = document.createElement("span")
        badge.classList.add("notification-badge")
        div.appendChild(badge)
      }

      badge.textContent = unreadCounts.get(fromUser)
    }
  }
}
