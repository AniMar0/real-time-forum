import { showSection } from './app.js';

const unreadCounts = new Map();
const chatCache = new Map(); // Cache messages per user
const messageOffsets = new Map(); // Track how many messages we've loaded per user
let socket = null;
let selectedUser = null;
let currentUser = null;
let isLoadingMessages = false; // Prevent multiple simultaneous loads

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

function renderMessage(msg, prepend = false) {
  const container = document.getElementById("chatMessages");
  const div = document.createElement("div");
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

// Throttle function to limit how often scroll handler executes
function throttle(func, delay) {
  let timeoutId;
  let lastExecTime = 0;
  return function (...args) {
    const currentTime = Date.now();
    
    if (currentTime - lastExecTime > delay) {
      func.apply(this, args);
      lastExecTime = currentTime;
    } else {
      clearTimeout(timeoutId);
      timeoutId = setTimeout(() => {
        func.apply(this, args);
        lastExecTime = Date.now();
      }, delay - (currentTime - lastExecTime));
    }
  };
}

// Load more messages when scrolling to top
async function loadMoreMessages() {
  if (!selectedUser || isLoadingMessages) return;
  
  isLoadingMessages = true;
  const currentOffset = messageOffsets.get(selectedUser) || 10;
  
  try {
    const res = await fetch(`/messages?from=${currentUser}&to=${selectedUser}&offset=${currentOffset}&limit=10`);
    if (!res.ok) throw new Error("Failed to load more messages");
    
    const messages = await res.json();
    
    if (messages.length === 0) {
      // No more messages to load
      isLoadingMessages = false;
      return;
    }
    
    // Store scroll position to maintain it after adding messages
    const container = document.getElementById("chatMessages");
    const oldScrollHeight = container.scrollHeight;
    
    // Add messages to the beginning of the chat
    messages.forEach(msg => renderMessage(msg, true));
    
    // Update cache by prepending new messages
    const cached = chatCache.get(selectedUser) || [];
    chatCache.set(selectedUser, [...messages, ...cached]);
    
    // Update offset
    messageOffsets.set(selectedUser, currentOffset + messages.length);
    
    // Maintain scroll position
    const newScrollHeight = container.scrollHeight;
    container.scrollTop = newScrollHeight - oldScrollHeight;
    
  } catch (err) {
    console.error("Error loading more messages:", err);
  } finally {
    isLoadingMessages = false;
  }
}

// Throttled scroll handler
const throttledScrollHandler = throttle((e) => {
  const container = e.target;
  // Check if scrolled to top (with small threshold)
  if (container.scrollTop <= 10) {
    loadMoreMessages();
  }
}, 300); // 300ms throttle delay

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
      
      const messagesContainer = document.getElementById("chatMessages");
      messagesContainer.innerHTML = "";

      // Reset offset for this user
      messageOffsets.set(username, 10);

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
          
          // Remove scroll event listener
          messagesContainer.removeEventListener('scroll', throttledScrollHandler);
        };
      }

      // Add scroll event listener for infinite scroll
      messagesContainer.addEventListener('scroll', throttledScrollHandler);

      // Load from cache or fetch initial messages
      const cachedMessages = chatCache.get(username);
      if (cachedMessages) {
        // Show only the last 10 messages initially
        const recentMessages = cachedMessages.slice(-10);
        recentMessages.forEach(renderMessage);
      } else {
        try {
          const res = await fetch(`/messages?from=${currentUser}&to=${selectedUser}&offset=0&limit=10`);
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