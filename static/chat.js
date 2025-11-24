import { errorToast } from "./toast.js"

const notificationsCache = new Map()
let socket = null
let selectedUser = null
let currentUser = null

let chatPage = 0
let isFetching = false
let noMoreMessages = false
let chatContainer = null
let displayedMessagesCount = 0
const renderedMessageIds = new Set()

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

function createChatWindow() {
  const existingChat = document.getElementById("chatWindow")
  if (existingChat) return existingChat

  const chatWindow = document.createElement("div")
  chatWindow.id = "chatWindow"
  chatWindow.classList.add("hidden", "chat-box")

  // Chat header
  const chatHeader = document.createElement("div")
  chatHeader.classList.add("chat-header")

  const headerStrong = document.createElement("strong")
  headerStrong.textContent = "Chat with: "

  const chatWithName = document.createElement("span")
  chatWithName.id = "chatWithName"

  headerStrong.appendChild(chatWithName)

  const closeBtn = document.createElement("i")
  closeBtn.id = "closeChatBtn"
  closeBtn.classList.add("fa-solid", "fa-xmark")
  closeBtn.style.cursor = "pointer"

  chatHeader.appendChild(headerStrong)
  chatHeader.appendChild(closeBtn)

  // Chat loader
  const chatLoader = document.createElement("div")
  chatLoader.id = "chatLoader"
  chatLoader.classList.add("hidden")
  chatLoader.style.textAlign = "center"
  chatLoader.style.padding = "5px"

  const loaderIcon = document.createElement("i")
  loaderIcon.classList.add("fa", "fa-spinner", "fa-spin")
  chatLoader.appendChild(loaderIcon)
  chatLoader.append(" Loading more...")

  // Chat messages
  const chatMessages = document.createElement("div")
  chatMessages.id = "chatMessages"

  // Typing indicator
  const typingIndicator = document.createElement("div")
  typingIndicator.id = "typingIndicator"
  typingIndicator.classList.add("hidden")

  // Message input
  const messageInput = document.createElement("input")
  messageInput.id = "messageInput"
  messageInput.type = "text"
  messageInput.placeholder = "Type a message..."

  // Send button
  const sendBtn = document.createElement("button")
  sendBtn.id = "sendBtn"
  sendBtn.textContent = "Send"

  chatWindow.appendChild(chatHeader)
  chatWindow.appendChild(chatLoader)
  chatWindow.appendChild(chatMessages)
  chatWindow.appendChild(typingIndicator)
  chatWindow.appendChild(messageInput)
  chatWindow.appendChild(sendBtn)

  const mainContent = document.getElementById("mainContent")
  mainContent.appendChild(chatWindow)

  return chatWindow
}

function setUserList(users) {
  const list = document.getElementById("userList")
  list.innerHTML = ""
  users.forEach((username) => {
    if (username.nickname === currentUser) return

    const div = document.createElement("div")
    div.className = "user"
    div.style.display = "flex"
    div.style.flexDirection = "column"
    div.style.gap = "8px"
    div.style.position = "relative"
    div.style.cursor = "pointer"
    div.style.padding = "12px"
    div.style.borderBottom = "1px solid #475569"
    div.style.transition = "all 0.2s"

    const mainRow = document.createElement("div")
    mainRow.style.display = "flex"
    mainRow.style.justifyContent = "space-between"
    mainRow.style.alignItems = "center"
    mainRow.style.width = "100%"

    const leftContainer = document.createElement("div")
    leftContainer.style.display = "flex"
    leftContainer.style.alignItems = "center"
    leftContainer.style.gap = "8px"

    const nameSpan = document.createElement("span")
    nameSpan.textContent = username.nickname
    nameSpan.style.fontWeight = "500"
    nameSpan.style.fontSize = "14px"

    const statusSpan = document.createElement("span")
    statusSpan.textContent = username.status
    statusSpan.style.fontSize = "12px"
    statusSpan.style.fontWeight = "500"
    statusSpan.style.padding = "2px 8px"
    statusSpan.style.borderRadius = "12px"

    if (username.status === "online") {
      statusSpan.style.color = "#4CAF50"
      statusSpan.style.backgroundColor = "rgba(76, 175, 80, 0.1)"
    } else {
      statusSpan.style.color = "#f44336"
      statusSpan.style.backgroundColor = "rgba(244, 67, 54, 0.1)"
    }

    leftContainer.appendChild(nameSpan)
    leftContainer.appendChild(statusSpan)

    const notifCount = notificationsCache.get(username.nickname) || 0
    if (notifCount > 0) {
      const badge = document.createElement("span")
      badge.classList.add("notification-badge")
      badge.textContent = notifCount
      badge.style.position = "static"
      badge.style.marginLeft = "auto"
      mainRow.appendChild(leftContainer)
      mainRow.appendChild(badge)
    } else {
      mainRow.appendChild(leftContainer)
    }

    div.appendChild(mainRow)

    const typingDiv = document.createElement("div")
    typingDiv.className = "typing-indicator-user"
    typingDiv.setAttribute("data-user", username.nickname)
    typingDiv.style.display = "none"
    typingDiv.style.fontSize = "12px"
    typingDiv.style.color = "#94a3b8"
    typingDiv.style.fontStyle = "italic"
    typingDiv.style.paddingLeft = "4px"
    typingDiv.innerHTML = '<span class="typing-dots">typing...</span>'

    div.appendChild(typingDiv)

    div.addEventListener("click", async () => {
      createChatWindow()

      const typingIndicator = document.getElementById("typingIndicator")

      typingIndicator.textContent = ""
      typingIndicator.classList.add("hidden")
      if (typingTimeoutSideBarId) clearTimeout(typingTimeoutSideBarId)

      chatPage = 0
      noMoreMessages = false
      displayedMessagesCount = 0
      renderedMessageIds.clear()
      chatContainer = document.getElementById("chatMessages")

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

      const closeChatBtn = document.getElementById("closeChatBtn")
      if (closeChatBtn) {
        closeChatBtn.onclick = () => {
          document.getElementById("chatWindow").classList.add("hidden")
          selectedUser = null
          document.getElementById("chatWithName").textContent = ""
          chatPage = 0
          noMoreMessages = false
          displayedMessagesCount = 0
          renderedMessageIds.clear()
        }
      }

      notificationsCache.set(username.nickname, 0)
      await markNotificationsAsRead(username.nickname)
      updateNotificationBadgeFromCache(username.nickname)

      try {
        const res = await fetch(`/messages?from=${currentUser}&to=${selectedUser}&offset=0`, {
          method: "POST",
        })
        if (!res.ok) throw new Error("Failed to load chat history")
        const messages = await res.json()

        if (messages && Array.isArray(messages)) {
          messages.forEach(renderMessage)
          displayedMessagesCount = messages.length
        }
      } catch (err) {
        console.error("Error loading chat history:", err)
        errorToast("Failed to load chat history")
      }
    })

    list.appendChild(div)
  })
}

export async function startChatFeature(currentUsername) {
  currentUser = currentUsername

  await loadNotificationsFromDB()

  socket = new WebSocket("ws://" + window.location.host + "/ws")

  socket.addEventListener("message", (event) => {
    const data = JSON.parse(event.data)

    if (data.event === "logout") {
      window.location.reload()
      return
    }

    if (data.type === "user_list") {
      console.log("Received user list update")
      setUserList(data.users)
    }

    if (data.type === "typing_indicator") {
      if (data.from === selectedUser && data.to === currentUser) {
        renderTypingIndicatorChatBox(data.from)
      } else if (data.to === currentUser) {
        renderTypingIndicatorSideBar(data.from)
      }
    }

    const messageId = getMessageId(data)

    if (renderedMessageIds.has(messageId)) return

    if (data.type === "chat_message") {
      if (data.from === selectedUser || data.to === selectedUser) {
        renderMessage(data)
        displayedMessagesCount++
        if (data.from === selectedUser) {
          notificationsCache.set(data.from, 0)
          markNotificationsAsRead(data.from)
          updateNotificationBadgeFromCache(data.from)
        }
      } else if (data.to === currentUser) {
        const currentCount = notificationsCache.get(data.from) || 0
        notificationsCache.set(data.from, currentCount + 1)
        updateNotificationBadgeFromCache(data.from)
      }
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
        type: "chat_message",
      }

      const messageId = getMessageId(message)
      if (renderedMessageIds.has(messageId)) return

      fetch("/sendMessage", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(message),
      })
        .then((res) => {
          if (!res.ok) throw new Error("Failed to send message")
          return res.json()
        })
        .then((message) => {
          socket.send(JSON.stringify(message))
          renderMessage(message)
          displayedMessagesCount++
        })
        .catch((err) => {
          errorToast("Failed to send message. Please try again.")
        })
      input.value = ""
    }
    sendBtn.addEventListener("click", sendMessage)
    input.addEventListener("keypress", (e) => {
      if (e.key === "Enter") {
        sendMessage()
      } else {
        const message = {
          to: selectedUser,
          from: currentUser,
          content: "",
          timestamp: "",
          type: "typing_indicator",
        }
        socket.send(JSON.stringify(message))
      }
    })
  }
}

async function loadMessagesPage(from, to) {
  if (isFetching || noMoreMessages) return

  isFetching = true
  const offset = displayedMessagesCount
  const loader = document.getElementById("chatLoader")
  const minDisplayTime = 500
  const start = Date.now()
  if (loader) loader.classList.remove("hidden")

  try {
    const res = await fetch(`/messages?from=${from}&to=${to}&offset=${offset}`, {
      method: "POST",
    })
    if (!res.ok) throw new Error("Failed to load chat messages")
    const messages = await res.json()

    if (!Array.isArray(messages) || messages.length === 0) {
      noMoreMessages = true
    } else {
      const container = document.getElementById("chatMessages")
      const oldScrollHeight = container.scrollHeight
      const oldScrollTop = container.scrollTop

      const newMessages = messages.filter((msg) => !renderedMessageIds.has(getMessageId(msg)))

      if (newMessages.length > 0) {
        const sortedMessages = newMessages
        sortedMessages.reverse().forEach((msg) => renderMessageAtTop(msg))

        displayedMessagesCount += newMessages.length

        const newScrollHeight = container.scrollHeight
        const heightDifference = newScrollHeight - oldScrollHeight
        container.scrollTop = oldScrollTop + heightDifference
      } else {
        noMoreMessages = true
      }
    }
  } catch (err) {
  } finally {
    const timeElapsed = Date.now() - start
    const remainingTime = minDisplayTime - timeElapsed

    setTimeout(
      () => {
        if (loader) loader.classList.add("hidden")
        isFetching = false
      },
      remainingTime > 0 ? remainingTime : 0,
    )
  }
}

const renderMessageAtTop = (msg) => {
  const messageId = getMessageId(msg)
  if (renderedMessageIds.has(messageId)) return

  const container = document.getElementById("chatMessages")
  const div = document.createElement("div")
  div.setAttribute("data-message-id", messageId)
  div.innerHTML = `
    <p><strong>${msg.from}</strong>: ${msg.content}<br/>
    <small>${new Date(msg.timestamp).toLocaleTimeString()}</small></p>
  `
  container.insertBefore(div, container.firstChild)
  renderedMessageIds.add(messageId)
}

function renderMessage(msg) {
  const messageId = getMessageId(msg)
  if (renderedMessageIds.has(messageId)) return

  const container = document.getElementById("chatMessages")

  const div = document.createElement("div")
  div.setAttribute("data-message-id", messageId)

  const p = document.createElement("p")

  const strong = document.createElement("strong")
  strong.textContent = msg.from

  p.appendChild(strong)
  p.append(": " + msg.content)

  p.appendChild(document.createElement("br"))

  const small = document.createElement("small")
  small.textContent = new Date(msg.timestamp).toLocaleTimeString()
  p.appendChild(small)

  div.appendChild(p)
  container.appendChild(div)

  container.scrollTop = container.scrollHeight
  renderedMessageIds.add(messageId)
}

function getMessageId(msg) {
  return msg.id || `${msg.timestamp}_${msg.from}_${msg.to}_${msg.content}`
}

async function loadNotificationsFromDB() {
  try {
    const res = await fetch("/notifications", {
      method: "GET",
    })
    if (!res.ok) {
      console.error("Failed to load notifications")
      return
    }
    const notifications = await res.json()

    notificationsCache.clear()
    for (const [sender, count] of Object.entries(notifications)) {
      notificationsCache.set(sender, count)
    }

    console.log("Notifications loaded from DB:", notificationsCache)
  } catch (err) {
    console.error("Error loading notifications:", err)
  }
}

function updateNotificationBadgeFromCache(username) {
  const userList = document.getElementById("userList")
  if (!userList) return

  const users = userList.querySelectorAll(".user")
  for (const userDiv of users) {
    const mainRow = userDiv.querySelector("div:first-child")
    if (!mainRow) continue

    const leftContainer = mainRow.querySelector("div:first-child")
    if (!leftContainer) continue

    const nameSpan = leftContainer.querySelector("span:first-child")
    if (nameSpan && nameSpan.textContent === username) {
      const count = notificationsCache.get(username) || 0
      let badge = mainRow.querySelector(".notification-badge")

      if (count > 0) {
        if (!badge) {
          badge = document.createElement("span")
          badge.classList.add("notification-badge")
          badge.style.position = "static"
          badge.style.marginLeft = "auto"
          mainRow.appendChild(badge)
        }
        badge.textContent = count
      } else if (badge) {
        badge.remove()
      }
      break
    }
  }
}

async function markNotificationsAsRead(sender) {
  try {
    const res = await fetch("/notifications/mark-read", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ sender: sender }),
    })
    if (!res.ok) {
      console.error("Failed to mark notifications as read")
    }
  } catch (err) {
    console.error("Error marking notifications as read:", err)
  }
}

let typingTimeoutSideBarId = null

function renderTypingIndicatorSideBar(from) {
  const userList = document.getElementById("userList")
  if (!userList) return

  const typingDiv = userList.querySelector(`.typing-indicator-user[data-user="${from}"]`)
  if (!typingDiv) return

  typingDiv.style.display = "block"
  const parentUserDiv = typingDiv.parentElement
  parentUserDiv.addEventListener("click", () => {
    typingDiv.style.display = "none"
    typingTimeoutSideBarId = null
  })
  if (typingTimeoutSideBarId) clearTimeout(typingTimeoutSideBarId)
  typingTimeoutSideBarId = setTimeout(() => {
    typingDiv.style.display = "none"
    typingTimeoutSideBarId = null
  }, 1000)
}

let typingTimeoutChatBoxId = null
function renderTypingIndicatorChatBox(from) {
  const typingDiv = document.getElementById("typingIndicator")
  if (!typingDiv) return

  typingDiv.classList.remove("hidden")
  typingDiv.textContent = `${from} is typing...`

  if (typingTimeoutChatBoxId) clearTimeout(typingTimeoutChatBoxId)
  typingTimeoutChatBoxId = setTimeout(() => {
    typingDiv.textContent = ""
    typingDiv.classList.add("hidden")
    typingTimeoutChatBoxId = null
  }, 1000)
}
