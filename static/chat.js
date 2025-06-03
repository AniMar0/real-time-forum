import { showSection } from './app.js';

let socket = null
let selectedUser = null

export function startChatFeature(currentUser) {

  socket = new WebSocket("ws://" + window.location.host + "/ws")

  socket.addEventListener("open", () => {
    console.log("✅ WebSocket connected")
  })

  socket.addEventListener("message", (event) => {
    const data = JSON.parse(event.data)

    if (data.type === "user_list") {
      setUserList(data.users, currentUser)
    } else {
      // استلام رسالة
      if (data.from === selectedUser || data.to === selectedUser) {
        renderMessage(data)
      }
    }
  })

  const form = document.getElementById("chatForm")
  if (form) {
    form.addEventListener("submit", (e) => {
      e.preventDefault()
      const input = document.getElementById("chatInput")
      const content = input.value.trim()
      if (!content || !selectedUser) return

      const message = {
        to: selectedUser,
        content: content,
      }

      socket.send(JSON.stringify(message))
      input.value = ""
    })
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
  container.scrollTop = container.scrollHeight
}

function setUserList(users, currentUser) {
  const list = document.getElementById("userList")
  list.innerHTML = ""
  users.forEach((userObj) => {

    const username = userObj
    const isOnline = true

    if (currentUser === userObj) return


    const div = document.createElement("div")
    div.className = "user"
    div.style.display = "flex"
    div.style.justifyContent = "space-between"
    div.style.alignItems = "center"
    div.style.cursor = "pointer"
    div.style.padding = "5px"
    div.style.borderBottom = "1px solid #ddd"

    const nameSpan = document.createElement("span")
    nameSpan.textContent = username

    const statusSpan = document.createElement("span")
    statusSpan.classList.add("status")
    if (isOnline) statusSpan.classList.add("online")

    div.appendChild(nameSpan)
    div.appendChild(statusSpan)

    div.addEventListener("click", async () => {
      selectedUser = username

      showSection("chatWindow")

      document.getElementById("chatMessages").innerHTML = ""

      try {
        const res = await fetch(`/messages?from=${username}&to=${selectedUser}`)
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
