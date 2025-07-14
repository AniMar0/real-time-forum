import { showSection } from './app.js'


const unreadCounts = new Map()
let socket = null
let selectedUser = null
let currentUser = null

export function startChatFeature(currentUsername) {
  currentUser = currentUsername

  socket = new WebSocket("ws://" + window.location.host + "/ws")

  socket.addEventListener("message", (event) => {
    const data = JSON.parse(event.data);

    if (data.type === "user_list") {
      setUserList(data.users);
    } else {
      if (data.from === selectedUser || data.to === selectedUser) {
        renderMessage(data);
      } else if (data.to === currentUser) {
        // Increase unread count
        const prev = unreadCounts.get(data.from) || 0;
        unreadCounts.set(data.from, prev + 1);
        updateNotificationBadge(data.from);
      }
    }
  });


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

function setUserList(users) {
  const list = document.getElementById("userList")
  list.innerHTML = ""

  users.forEach((username) => {
    if (username === currentUser) return

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
    statusSpan.classList.add("status", "online")

    div.appendChild(nameSpan)
    div.appendChild(statusSpan)

    div.addEventListener("click", async () => {
      selectedUser = username
      document.getElementById("chatWithName").textContent = username;
      document.getElementById("chatWindow").classList.remove("hidden")
      document.getElementById("chatMessages").innerHTML = ""

      unreadCounts.set(username, 0);
      const badge = div.querySelector(".notification-badge");
      if (badge) badge.remove();

      const closeChatBtn = document.getElementById("closeChatBtn");
      if (closeChatBtn) {
        closeChatBtn.onclick = () => {
          document.getElementById("chatWindow").classList.add("hidden");
          selectedUser = null;
          document.getElementById("chatMessages").innerHTML = "";
          document.getElementById("chatWithName").textContent = "";
        };
      }
F

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


//function showNotificationBadge(fromUser) {
//const userList = document.getElementById("userList");
// const users = userList.getElementsByClassName("user");

//for (let div of users) {
//const nameSpan = div.querySelector("span:first-child");
//if (nameSpan && nameSpan.textContent === fromUser) {
// Prevent adding multiple badges
//if (!div.querySelector(".notification-badge")) {
//const badge = document.createElement("span");
//badge.classList.add("notification-badge");
//badge.textContent = "•";
//div.appendChild(badge);
//}
//}
//}
//}


function updateNotificationBadge(fromUser) {
  const userList = document.getElementById("userList");
  const users = userList.getElementsByClassName("user");

  for (let div of users) {
    const nameSpan = div.querySelector("span:first-child");
    if (nameSpan && nameSpan.textContent === fromUser) {
      let badge = div.querySelector(".notification-badge");

      // Create badge if it doesn't exist
      if (!badge) {
        badge = document.createElement("span");
        badge.classList.add("notification-badge");
        div.appendChild(badge);
      }

      badge.textContent = unreadCounts.get(fromUser);
    }
  }
}

