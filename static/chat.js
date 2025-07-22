import { showSection } from './app.js';

const unreadCounts = new Map();
const chatCache = new Map(); // Cache messages per user
const messageOffsets = new Map(); // Track how many messages loaded per user
let socket = null;
let selectedUser = null;
let currentUser = null;
let isLoadingMessages = false;

export function startChatFeature(currentUsername) {
  currentUser = currentUsername;

  socket = new WebSocket("ws://" + window.location.host + "/ws");

  socket.addEventListener("message", (event) => {
    const data = JSON.parse(event.data);

    if (data.type === "user_list") {
      setUserList(data.users);
    } else {
      if (data.from === selectedUser || data.to === selectedUser) {
        renderMessage(data, false); // false = append to bottom

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
      renderMessage(message, false); // false = append to bottom

      // Update cache
      const cached = chatCache.get(selectedUser) || [];
      chatCache.set(selectedUser, [...cached, message]);

      input.value = "";
    });

    // Allow sending with Enter key
    input.addEventListener("keypress", (e) => {
      if (e.key === "Enter") {
        sendBtn.click();
      }
    });
  }
}

function renderMessage(msg, prepend = false) {
  const container = document.getElementById("chatMessages");
  const div = document.createElement("div");
  div.className = "message";
  div.innerHTML = `
    <p><strong>${msg.from}</strong>: ${msg.content}<br/>
    <small>${new Date(msg.timestamp).toLocaleTimeString()}</small></p>
  `;

  if (prepend) {
    container.insertBefore(div, container.firstChild);
  } else {
    container.appendChild(div);
    container.scrollTop = container.scrollHeight;
  }
}

// Throttle function to limit how often a function can be called
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
  }
}

// Load more messages from server
async function loadMoreMessages(username, offset) {
  if (isLoadingMessages) return;
  
  isLoadingMessages = true;
  
  try {
    const res = await fetch(`/messages?from=${currentUser}&to=${username}&offset=${offset}&limit=10`);
    if (!res.ok) throw new Error("Failed to load more messages");
    
    const messages = await res.json();
    
    if (messages.length > 0) {
      const chatMessages = document.getElementById("chatMessages");
      const scrollHeight = chatMessages.scrollHeight;
      
      // Prepend messages to the top
      messages.reverse().forEach(msg => renderMessage(msg, true));
      
      // Update offset
      const currentOffset = messageOffsets.get(username) || 10;
      messageOffsets.set(username, currentOffset + messages.length);
      
      // Update cache
      const cached = chatCache.get(username) || [];
      chatCache.set(username, [...messages, ...cached]);
      
      // Maintain scroll position
      chatMessages.scrollTop = chatMessages.scrollHeight - scrollHeight;
    }
  } catch (err) {
    console.error("Error loading more messages:", err);
  } finally {
    isLoadingMessages = false;
  }
}

// Throttled scroll handler
const handleScroll = throttle((e) => {
  const container = e.target;
  
  // Check if scrolled to top (with small buffer)
  if (container.scrollTop < 50 && selectedUser && !isLoadingMessages) {
    const currentOffset = messageOffsets.get(selectedUser) || 10;
    loadMoreMessages(selectedUser, currentOffset);
  }
}, 200); // Throttle to 200ms

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
      const chatMessages = document.getElementById("chatMessages");
      chatMessages.innerHTML = "";

      // Reset offset for this user
      messageOffsets.set(username, 10);

      unreadCounts.set(username, 0);
      const badge = div.querySelector(".notification-badge");
      if (badge) badge.remove();

      // Setup scroll event listener for infinite scrolling
      chatMessages.removeEventListener("scroll", handleScroll); // Remove previous listener
      chatMessages.addEventListener("scroll", handleScroll);

      // Close chat button setup
      const closeChatBtn = document.getElementById("closeChatBtn");
      if (closeChatBtn) {
        closeChatBtn.onclick = () => {
          document.getElementById("chatWindow").classList.add("hidden");
          selectedUser = null;
          document.getElementById("chatWithName").textContent = "";
          chatMessages.removeEventListener("scroll", handleScroll);
        };
      }

      // Load initial messages (last 10)
      try {
        const res = await fetch(`/messages?from=${currentUser}&to=${selectedUser}&limit=10`);
        if (!res.ok) throw new Error("Failed to load chat history");
        const messages = await res.json();
        
        // Clear and update cache
        chatCache.set(selectedUser, messages);
        
        messages.forEach(msg => renderMessage(msg, false));
      } catch (err) {
        console.error("Chat history error:", err);
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