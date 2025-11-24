import { showSection, logged } from "./app.js"
import { startChatFeature } from "./chat.js"
import { createErrorPage } from "./error.js"

export function createLoginSection() {
  const section = document.createElement("section")
  section.id = "loginSection"

  const h2 = document.createElement("h2")
  h2.textContent = "Login"

  const form = document.createElement("form")
  form.id = "loginForm"

  const identifierInput = document.createElement("input")
  identifierInput.id = "identifier"
  identifierInput.placeholder = "Email or Nickname"
  identifierInput.required = true

  const passwordInput = document.createElement("input")
  passwordInput.id = "loginPassword"
  passwordInput.placeholder = "Password"
  passwordInput.type = "password"
  passwordInput.required = true

  const errorDiv = document.createElement("div")
  errorDiv.id = "loginError"
  errorDiv.style.color = "red"
  errorDiv.style.display = "none"
  errorDiv.style.marginBottom = "10px"

  const submitBtn = document.createElement("button")
  submitBtn.type = "submit"
  submitBtn.textContent = "Login"

  form.appendChild(identifierInput)
  form.appendChild(passwordInput)
  form.appendChild(errorDiv)
  form.appendChild(submitBtn)

  form.addEventListener("submit", handleLogin)

  section.appendChild(h2)
  section.appendChild(form)

  return section
}

export function handleLogin(event) {
  event.preventDefault()
  const loginError = document.getElementById("loginError")
  loginError.style.display = "none"

  const formData = {
    identifier: document.getElementById("identifier").value,
    password: document.getElementById("loginPassword").value,
  }

  fetch("/login", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(formData),
  })
    .then((res) => {
      if (res.status != 200 && res.status != 401 && res.status != 201) {
        createErrorPage(res)
      }
      if (!res.ok) {
        throw new Error("Registration failed")
      }
      return res.json()
    })
    .then((data) => {
      startChatFeature(data.username)
      showSection("posts")
      logged(true, data.username)
    })
    .catch((err) => {
      loginError.textContent = "Incorrect email/nickname or password."
      loginError.style.display = "block"
      logged(false)
    })
}
