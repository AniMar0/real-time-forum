import { showSection } from './app.js';

const unreadCounts = new Map()
const chatCache = new Map() // Cache messages per user
let socket = null
let selectedUser = null
let currentUser = null
let isLoadingMessages = false
const messageOffsets = new Map()

export function startChatFeature(currentUsername) {
  currentUser = currentUsername;

  socket = new WebSocket("ws://" + window.location.host + "/ws");

  socket.addEventListener("message", (event) => {
    const data = JSON.parse(event.data);

    if (data.type === "user_list") {
      setUserList(data.users);
    } else {
      if (data.from === selectedUser || data.to === selectedUser) {
        renderMessage(data);

        // Also update cache
        const chatKey = data.from === currentUser ? data.to : data.from;
        const cached = chatCache.get(chatKey) || [];
        chatCache.set(chatKey, [...cached, data]);
      } else if (data.to === currentUser) {
        const prev = unreadCounts.get(data.from) || 0;
        unreadCounts.set(data.from, prev + 1);
        updateNotificationBadge(data.from);
      }
    }
  });

  const sendBtn = document.getElementById("sendBtn");
  const input = document.getElementById("messageInput");

  if (sendBtn && input) {
    sendBtn.addEventListener("click", () => {
      
      const content = input.value.trim();
      if (!content || !selectedUser) return;

      const message = {
        to: selectedUser,
        from: currentUser,
        content: content,
        timestamp: new Date().toISOString(),
      };

      socket.send(JSON.stringify(message));
      renderMessage(message);

      // Update cache
      const cached = chatCache.get(selectedUser) || [];
      chatCache.set(selectedUser, [...cached, message]);

      input.value = "";
    });
  }
}

const renderMessage = (msg) => {
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
  const list = document.getElementById("userList");
  list.innerHTML = "";

  users.forEach((username) => {
    if (username === currentUser) return;

    const div = document.createElement("div");
    div.className = "user";
    div.style.display = "flex";
    div.style.justifyContent = "space-between";
    div.style.alignItems = "center";
    div.style.cursor = "pointer";
    div.style.padding = "5px";
    div.style.borderBottom = "1px solid #ddd";

    const nameSpan = document.createElement("span");
    nameSpan.textContent = username;

    const statusSpan = document.createElement("span");
    statusSpan.classList.add("status", "online");

    div.appendChild(nameSpan);
    div.appendChild(statusSpan);

    div.addEventListener("click", async () => {
      selectedUser = username;
      document.getElementById("chatWithName").textContent = username;
      document.getElementById("chatWindow").classList.remove("hidden");
      document.getElementById("chatMessages").innerHTML = "";

      unreadCounts.set(username, 0);
      const badge = div.querySelector(".notification-badge");
      if (badge) badge.remove();

      // Close chat button setup
      const closeChatBtn = document.getElementById("closeChatBtn");
      if (closeChatBtn) {
        closeChatBtn.onclick = () => {
          document.getElementById("chatWindow").classList.add("hidden");
          selectedUser = null;
          document.getElementById("chatWithName").textContent = "";
        };
      }

      // Load from cache or fetch
      const cachedMessages = chatCache.get(username);
      if (cachedMessages) {
        cachedMessages.forEach(renderMessage);
      } else {
        try {
          const res = await fetch(`/messages?from=${currentUser}&to=${selectedUser}`);
          if (!res.ok) throw new Error("Failed to load chat history");
          const messages = await res.json();
          chatCache.set(selectedUser, messages);
          messages.forEach(renderMessage);
        } catch (err) {
          console.error("Chat history error:", err);
        }
      }
    });

    list.appendChild(div);
  });
}

function updateNotificationBadge(fromUser) {
  const userList = document.getElementById("userList");
  const users = userList.getElementsByClassName("user");

  for (let div of users) {
    const nameSpan = div.querySelector("span:first-child");
    if (nameSpan && nameSpan.textContent === fromUser) {
      let badge = div.querySelector(".notification-badge");

      if (!badge) {
        badge = document.createElement("span");
        badge.classList.add("notification-badge");
        div.appendChild(badge);
      }

      badge.textContent = unreadCounts.get(fromUser);
    }
  }
}
