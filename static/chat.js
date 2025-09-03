import { logged, showSection } from './app.js';
import { ErrorPage } from './error.js';

const unreadCounts = new Map()
const chatCache = new Map()
let socket = null
let selectedUser = null
let currentUser = null

let chatPage = 0
const messagePerPage = 10
let isFetching = false
let noMoreMessages = false
let chatContainer = null
let displayedMessagesCount = 0
let renderedMessageIds = new Set() // Track rendered messages to prevent duplicates

const throttle = (fn, wait) => {
  let lastTime = 0
  return function (...args) {
    const now = new Date().getTime()
    if (now - lastTime >= wait) {
      lastTime = now
      fn.apply(this, args)
    }
  }
}

// Generate unique ID for messages
function getMessageId(msg) {
  return msg.id || `${msg.timestamp}_${msg.from}_${msg.to}_${msg.content}`
}

async function loadMessagesPage(from, to, page) {
  if (isFetching || noMoreMessages) return // Prevent concurrent requests
  
  isFetching = true
  const offset = displayedMessagesCount
  const loader = document.getElementById("chatLoader")
  const minDisplayTime = 500
  const start = Date.now()
  if (loader) loader.classList.remove("hidden")

  try {
    const res = await fetch(`/messages?from=${from}&to=${to}&offset=${offset}`, {
      method: "POST"
    })
    if (!res.ok) throw new Error("Failed to load chat messages")
    const messages = await res.json()
    
    if (!Array.isArray(messages) || messages.length === 0) {
      noMoreMessages = true
    } else {
      const container = document.getElementById("chatMessages")
      const oldScrollHeight = container.scrollHeight
      const oldScrollTop = container.scrollTop
      
      // Filter out messages that are already rendered
      const newMessages = messages.filter(msg => !renderedMessageIds.has(getMessageId(msg)))
      
      if (newMessages.length > 0) {
        const sortedMessages = newMessages.sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp))
        sortedMessages.reverse().forEach(msg => renderMessageAtTop(msg))

        displayedMessagesCount += newMessages.length

        const newScrollHeight = container.scrollHeight
        const heightDifference = newScrollHeight - oldScrollHeight
        container.scrollTop = oldScrollTop + heightDifference

        // Update cache with new messages
        const cached = chatCache.get(to) || []
        const chronologicalMessages = newMessages.sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp))
        const mergedCache = mergeMessages([...chronologicalMessages, ...cached], [])
        chatCache.set(to, mergedCache)
      } else {
        // All messages were duplicates, consider this as no more messages
        noMoreMessages = true
      }
    }
  } catch (err) {
    console.error("Pagination error:", err)
  } finally {
    const timeElapsed = Date.now() - start
    const remainingTime = minDisplayTime - timeElapsed

    setTimeout(() => {
      if (loader) loader.classList.add("hidden")
      isFetching = false
    }, remainingTime > 0 ? remainingTime : 0)
  }
}

const renderMessageAtTop = (msg) => {
  const messageId = getMessageId(msg)
  if (renderedMessageIds.has(messageId)) return // Skip if already rendered
  
  const container = document.getElementById("chatMessages")
  const div = document.createElement("div")
  div.setAttribute('data-message-id', messageId) // Add ID to DOM element
  div.innerHTML = `
    <p><strong>${msg.from}</strong>: ${msg.content}<br/>
    <small>${new Date(msg.timestamp).toLocaleTimeString()}</small></p>
  `
  container.insertBefore(div, container.firstChild)
  renderedMessageIds.add(messageId)
}

export function startChatFeature(currentUsername) {
  currentUser = currentUsername
  socket = new WebSocket("ws://" + window.location.host + "/ws")

  socket.addEventListener("message", (event) => {
    const data = JSON.parse(event.data)

    if (data.event === "logout") {
      window.location.reload()
      return
    }

    if (data.type === "user_list") {
      setUserList(data.users)
      return
    }

    const chatKey = data.from === currentUser ? data.to : data.from
    const messageId = getMessageId(data)

    // Check if message is already rendered to prevent real-time duplicates
    if (renderedMessageIds.has(messageId)) return

    if (data.from === selectedUser || data.to === selectedUser) {
      renderMessage(data)
      displayedMessagesCount++
      const cached = chatCache.get(chatKey) || []
      const mergedCache = mergeMessages(cached, [data])
      chatCache.set(chatKey, mergedCache)
    } else if (data.to === currentUser) {
      const cached = chatCache.get(chatKey) || []
      const mergedCache = mergeMessages(cached, [data])
      chatCache.set(chatKey, mergedCache)
      notification(data.to, data.from, 1)
    }
  })

  const sendBtn = document.getElementById("sendBtn")
  const input = document.getElementById("messageInput")
  if (sendBtn && input) {
    const sendMessage = () => {
      const content = input.value.trim()
      if (!content || !selectedUser) return

      const message = {
        to: selectedUser,
        from: currentUser,
        content: content,
        timestamp: new Date().toISOString(),
      }
      
      const messageId = getMessageId(message)
      if (renderedMessageIds.has(messageId)) return // Prevent duplicate sends
      
      socket.send(JSON.stringify(message))
      renderMessage(message)
      displayedMessagesCount++
      const cached = chatCache.get(selectedUser) || []
      const mergedCache = mergeMessages(cached, [message])
      chatCache.set(selectedUser, mergedCache)
      input.value = ""
    }
    sendBtn.addEventListener("click", sendMessage)
  }
}

function renderMessage(msg) {
  const messageId = getMessageId(msg)
  if (renderedMessageIds.has(messageId)) return // Skip if already rendered
  
  const container = document.getElementById("chatMessages")
  const div = document.createElement("div")
  div.setAttribute('data-message-id', messageId)
  div.innerHTML = `
    <p><strong>${msg.from}</strong>: ${msg.content}<br/>
    <small>${new Date(msg.timestamp).toLocaleTimeString()}</small></p>
  `
  container.appendChild(div)
  container.scrollTop = container.scrollHeight
  renderedMessageIds.add(messageId)
}

function setUserList(users) {
  const list = document.getElementById("userList")
  list.innerHTML = ""
  users.forEach((username) => {
    if (username.nickname === currentUser) return

    const div = document.createElement("div")
    div.className = "user"
    div.style.display = "flex"
    div.style.justifyContent = "space-between"
    div.style.alignItems = "center"
    div.style.cursor = "pointer"
    div.style.padding = "5px"
    div.style.borderBottom = "1px solid #ddd"
    const nameSpan = document.createElement("span")
    nameSpan.textContent = username.nickname
    const statusSpan = document.createElement("span")
    statusSpan.classList.add("status", username.status)
    div.appendChild(nameSpan)
    div.appendChild(statusSpan)
    notification(currentUser, username.nickname)
    
    div.addEventListener("click", async () => {
      // Reset all pagination and rendering state for new chat
      chatPage = 0
      noMoreMessages = false
      displayedMessagesCount = 0
      renderedMessageIds.clear() // Clear rendered message tracking
      chatContainer = document.getElementById("chatMessages")

      // Remove existing scroll handler
      const existingHandler = chatContainer.scrollHandler
      if (existingHandler) {
        chatContainer.removeEventListener("scroll", existingHandler)
      }

      const scrollHandler = throttle(async () => {
        const isNearTop = chatContainer.scrollTop <= 100
        const isAtTop = chatContainer.scrollTop === 0

        if ((isNearTop || isAtTop) && !isFetching && !noMoreMessages) {
          chatPage += 1
          await loadMessagesPage(currentUser, selectedUser, chatPage)
        }
      }, 200)
      chatContainer.scrollHandler = scrollHandler
      chatContainer.addEventListener("scroll", scrollHandler)

      selectedUser = username.nickname
      document.getElementById("chatWithName").textContent = username.nickname
      document.getElementById("chatWindow").classList.remove("hidden")
      document.getElementById("chatMessages").innerHTML = ""

      const badge = div.querySelector(".notification-badge")
      if (badge) badge.remove()

      const closeChatBtn = document.getElementById("closeChatBtn")
      if (closeChatBtn) {
        closeChatBtn.onclick = () => {
          document.getElementById("chatWindow").classList.add("hidden")
          selectedUser = null
          document.getElementById("chatWithName").textContent = ""
          // Reset all state when closing chat
          chatPage = 0
          noMoreMessages = false
          displayedMessagesCount = 0
          renderedMessageIds.clear()
        }
      }
      notification(currentUser, username.nickname, 0)

      // Load initial messages
      try {
        const res = await fetch(`/messages?from=${currentUser}&to=${selectedUser}&offset=0`, {
          method: "POST"
        })
        if (!res.ok) throw new Error("Failed to load chat history")
        const messages = await res.json()

        const cached = chatCache.get(username.nickname) || []
        const merged = mergeMessages(cached, messages)
        chatCache.set(username.nickname, merged)

        // Render unique messages only
        const sortedMerged = merged.sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp))
        sortedMerged.forEach(renderMessage)
        displayedMessagesCount = sortedMerged.length
      } catch (err) {
        console.error("Chat history error:", err)
      }
    })

    list.appendChild(div)
  })
}

function mergeMessages(oldMessages, newMessages) {
  const all = [...oldMessages, ...newMessages]
  const seen = new Set()
  return all.filter(msg => {
    const id = getMessageId(msg)
    if (seen.has(id)) return false
    seen.add(id)
    return true
  })
}

function updateNotificationBadge(data) {
  const userList = document.getElementById("userList")
  const users = userList.getElementsByClassName("user")
  if (!userList || data.unread_messages == 0) return

  for (let div of users) {
    const nameSpan = div.querySelector("span:first-child")
    if (nameSpan && nameSpan.textContent === data.sender_nickname) {
      let badge = div.querySelector(".notification-badge")

      if (!badge) {
        badge = document.createElement("span")
        badge.classList.add("notification-badge")
        div.appendChild(badge)
      }
      badge.textContent = data.unread_messages
    }
  }
}

function notification(receiver, sender, unread) {
  const notifData = {
    receiver_nickname: receiver,
    sender_nickname: sender,
    ...(unread != null && { unread_messages: unread })
  }

  fetch("/notification", {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify(notifData)
  })
    .then(res => {
      if (res.status != 200 && res.status != 401 && res.status != 201) {
        ErrorPage(res)
      }
      if (!res.ok) {
        throw new Error("notif failed")
      }
      return res.json()
    })
    .then(data => {
      updateNotificationBadge(data)
    })
    .catch(err => {
      alert("Sesion determ")
      window.location.reload()
      console.error(err)
    })
}