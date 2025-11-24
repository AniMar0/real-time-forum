import { createRegisterSection } from "./register.js"
import { startChatFeature } from "./chat.js"
import { createLoginSection } from "./login.js"
import { loadPosts, createPostsSection } from "./posts.js"
import { logout } from "./logout.js"
import { createErrorPage } from "./error.js"

let currentSection = null

function createHeader() {
  const header = document.createElement("header")

  const h1 = document.createElement("h1")
  h1.textContent = "My Forum"

  const nav = document.createElement("nav")

  const usernameDisplay = document.createElement("span")
  usernameDisplay.id = "usernameDisplay"
  usernameDisplay.classList.add("hidden")

  const showLoginBtn = document.createElement("button")
  showLoginBtn.id = "showLogin"
  showLoginBtn.textContent = "Login"
  showLoginBtn.addEventListener("click", () => showSection("login"))

  const showRegisterBtn = document.createElement("button")
  showRegisterBtn.id = "showRegister"
  showRegisterBtn.textContent = "Register"
  showRegisterBtn.addEventListener("click", () => showSection("register"))

  const logoutBtn = document.createElement("button")
  logoutBtn.id = "logoutBtn"
  logoutBtn.textContent = "Logout"
  logoutBtn.classList.add("hidden")
  logoutBtn.addEventListener("click", (e) => {
    logout(e)
    logged(false)
  })

  nav.appendChild(usernameDisplay)
  nav.appendChild(showLoginBtn)
  nav.appendChild(showRegisterBtn)
  nav.appendChild(logoutBtn)

  header.appendChild(h1)
  header.appendChild(nav)

  return header
}

function createMainContainer() {
  const main = document.createElement("main")
  main.id = "my-content"
  main.classList.add("flex-container")

  const userList = document.createElement("div")
  userList.id = "userList"
  userList.classList.add("hidden")

  const mainContent = document.createElement("div")
  mainContent.classList.add("main-content")
  mainContent.id = "mainContent"

  main.appendChild(userList)
  main.appendChild(mainContent)

  return main
}

function initApp() {
  const app = document.getElementById("app")

  const header = createHeader()
  const main = createMainContainer()

  app.appendChild(header)
  app.appendChild(main)
}

const checkLoggedIn = () => {
  fetch("/logged", {
    method: "POST",
    credentials: "include",
  })
    .then((res) => {
      if (res.status != 200 && res.status != 401 && res.status != 405 && res.status != 201) {
        createErrorPage(res)
      }
      if (!res.ok) throw new Error("Not logged in")
      return res.json()
    })
    .then((data) => {
      logged(true, data.username)
      startChatFeature(data.username)
      showSection("posts")
    })
    .catch(() => {
      logged(false)
      showSection("login")
    })
}

setInterval(() => {
  checkLoggedIn()
}, 60000)

document.addEventListener("DOMContentLoaded", () => {
  initApp()

  if (window.location.pathname != "/") {
    createErrorPage({ status: 404, statusText: "Page not found" })
  }
  checkLoggedIn()
})

window.addEventListener("storage", (event) => {
  if (event.key === "logout") {
    window.location.reload()
  }
})

export function showSection(sectionId) {
  const mainContent = document.getElementById("mainContent")
  mainContent.innerHTML = "" // Clear previous section

  let sectionElement

  switch (sectionId) {
    case "login":
      sectionElement = createLoginSection()
      break
    case "register":
      sectionElement = createRegisterSection()
      break
    case "posts":
      sectionElement = createPostsSection()
      loadPosts()
      break
    default:
      console.error("Unknown section:", sectionId)
      return
  }

  if (sectionElement) {
    mainContent.appendChild(sectionElement)
    currentSection = sectionId
  }
}

export function logged(bool, user) {
  const usernameDisplay = document.getElementById("usernameDisplay")
  const showLoginBtn = document.getElementById("showLogin")
  const showRegisterBtn = document.getElementById("showRegister")
  const logoutBtn = document.getElementById("logoutBtn")
  const userList = document.getElementById("userList")

  if (bool) {
    usernameDisplay.textContent = user
    showLoginBtn.classList.add("hidden")
    showRegisterBtn.classList.add("hidden")
    logoutBtn.classList.remove("hidden")
    userList.classList.remove("hidden")
    usernameDisplay.classList.remove("hidden")
  } else {
    usernameDisplay.textContent = ""
    usernameDisplay.classList.add("hidden")
    showLoginBtn.classList.remove("hidden")
    showRegisterBtn.classList.remove("hidden")
    logoutBtn.classList.add("hidden")
    userList.classList.add("hidden")
  }
}
