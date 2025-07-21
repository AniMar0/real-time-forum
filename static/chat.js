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

const  renderMessage = (msg, prepend = false) => {
  const container = document.getElementById("chatMessages")
  const div = document.createElement("div")
  div.innerHTML = `
    <p><strong>${msg.from}</strong>: ${msg.content}<br/>
    <small>${new Date(msg.timestamp).toLocaleTimeString()}</small></p>
  `

  if (prepend) {
    container.insertBefore(div, container.firstChild)
  } else {
    container.appendChild(div)
    container.scrollTop = container.scrollHeight
  }
}

// Throttled scroll handler to prevent excessive API calls
function throttle(func, limit) {
  let inThrottle;
  return function() {
    const args = arguments;
    const context = this;
    if (!inThrottle) {
      func.apply(context, args);
      inThrottle = true;
      setTimeout(() => inThrottle = false, limit);
    }
  };
}

function setupScrollListener() {
  const chatMessages = document.getElementById("chatMessages");
  if (!chatMessages) return;

  const throttledScrollHandler = throttle(async function() {
    // Check if scrolled to top (with small threshold)
    if (chatMessages.scrollTop <= 10 && !isLoadingMessages && selectedUser) {
      await loadMoreMessages();
    }
  }, 300); // Throttle to 300ms

  chatMessages.addEventListener('scroll', throttledScrollHandler);
}

async function loadMoreMessages() {
  if (!selectedUser || isLoadingMessages) return;
  
  isLoadingMessages = true;
  const currentOffset = messageOffsets.get(selectedUser) || 0;
  
  try {
    const response = await fetch(`/messages?from=${currentUser}&to=${selectedUser}&offset=${currentOffset}&limit=10`);
    if (!response.ok) throw new Error("Failed to load more messages");
    
    const messages = await response.json();
    
    if (messages.length > 0) {
      // Store scroll position before adding messages
      const chatMessages = document.getElementById("chatMessages");
      const oldScrollHeight = chatMessages.scrollHeight;
      
      // Add messages to the beginning of the container
      messages.reverse().forEach(msg => renderMessage(msg, true));
      
      // Restore scroll position (maintain user's view)
      const newScrollHeight = chatMessages.scrollHeight;
      chatMessages.scrollTop = newScrollHeight - oldScrollHeight;
      
      // Update offset for next load
      messageOffsets.set(selectedUser, currentOffset + messages.length);
      
      // Update cache
      const cached = chatCache.get(selectedUser) || [];
      chatCache.set(selectedUser, [...messages, ...cached]);
    }
  } catch (error) {
    console.error("Error loading more messages:", error);
  } finally {
    isLoadingMessages = false;
  }
}

async function loadInitialMessages(username) {
  try {
    // Load last 10 messages
    const response = await fetch(`/messages?from=${currentUser}&to=${username}&offset=0&limit=10&order=desc`);
    if (!response.ok) throw new Error("Failed to load chat history");
    
    const messages = await response.json();
    
    if (messages.length > 0) {
      // Clear any existing messages
      document.getElementById("chatMessages").innerHTML = "";
      
      // Render messages in correct order (oldest first)
      messages.reverse().forEach(msg => renderMessage(msg));
      
      // Update offset and cache
      messageOffsets.set(username, messages.length);
      chatCache.set(username, messages);
      
      // Scroll to bottom
      const chatMessages = document.getElementById("chatMessages");
      chatMessages.scrollTop = chatMessages.scrollHeight;
    }
  } catch (error) {
    console.error("Error loading initial messages:", error);
  }
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
