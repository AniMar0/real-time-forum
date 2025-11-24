import { createErrorPage } from "./error.js"
import { errorToast } from "./toast.js"

export function createRegisterSection() {
  const section = document.createElement("section")
  section.id = "registerSection"

  const h2 = document.createElement("h2")
  h2.textContent = "Register"

  const form = document.createElement("form")
  form.id = "registerForm"

  const nicknameInput = document.createElement("input")
  nicknameInput.id = "nickname"
  nicknameInput.placeholder = "Nickname"
  nicknameInput.required = true

  const firstNameInput = document.createElement("input")
  firstNameInput.id = "firstName"
  firstNameInput.placeholder = "First Name"
  firstNameInput.required = true

  const lastNameInput = document.createElement("input")
  lastNameInput.id = "lastName"
  lastNameInput.placeholder = "Last Name"
  lastNameInput.required = true

  const emailInput = document.createElement("input")
  emailInput.id = "email"
  emailInput.placeholder = "Email"
  emailInput.type = "email"
  emailInput.required = true

  const passwordInput = document.createElement("input")
  passwordInput.id = "password"
  passwordInput.placeholder = "Password"
  passwordInput.type = "password"
  passwordInput.required = true

  const dobInput = document.createElement("input")
  dobInput.id = "dateOfBirth"
  dobInput.type = "date"
  dobInput.required = true

  const genderSelect = document.createElement("select")
  genderSelect.id = "gender"
  genderSelect.required = true

  const genderPlaceholder = document.createElement("option")
  genderPlaceholder.value = ""
  genderPlaceholder.textContent = "Select Gender"

  const maleOption = document.createElement("option")
  maleOption.value = "male"
  maleOption.textContent = "Male"

  const femaleOption = document.createElement("option")
  femaleOption.value = "female"
  femaleOption.textContent = "Female"

  genderSelect.appendChild(genderPlaceholder)
  genderSelect.appendChild(maleOption)
  genderSelect.appendChild(femaleOption)

  const submitBtn = document.createElement("button")
  submitBtn.type = "submit"
  submitBtn.textContent = "Register"

  form.appendChild(nicknameInput)
  form.appendChild(firstNameInput)
  form.appendChild(lastNameInput)
  form.appendChild(emailInput)
  form.appendChild(passwordInput)
  form.appendChild(dobInput)
  form.appendChild(genderSelect)
  form.appendChild(submitBtn)

  form.addEventListener("submit", handleRegister)

  section.appendChild(h2)
  section.appendChild(form)

  return section
}

export function handleRegister(event) {
  event.preventDefault()

  const dob = document.getElementById("dateOfBirth").value
  if (!dob) {
    errorToast("Your age is not accepted")
    return
  }
  const Age = calculateAge(dob)

  if (Age < 13 || Age > 120) {
    errorToast("Your age is not accepted. Must be between 13 and 120 years old.")
    return
  }

  if (document.getElementById("nickname").value == "") {
    errorToast("Nickname is required")
    return
  }
  if (document.getElementById("firstName").value == "") {
    errorToast("First name is required")
    return
  }
  if (document.getElementById("lastName").value == "") {
    errorToast("Last name is required")
    return
  }
  if (document.getElementById("email").value == "") {
    errorToast("Email is required")
    return
  }
  if (document.getElementById("password").value == "" || document.getElementById("password").value.length < 8) {
    errorToast("Password must be at least 8 characters long")
    return
  }
  if (document.getElementById("gender").value == "") {
    errorToast("Gender is required")
    return
  }

  const formData = {
    nickname: document.getElementById("nickname").value,
    first_name: document.getElementById("firstName").value,
    last_name: document.getElementById("lastName").value,
    email: document.getElementById("email").value,
    password: document.getElementById("password").value,
    age: Age,
    gender: document.getElementById("gender").value,
  }

  fetch("/register", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(formData),
  })
    .then(async (res) => {
      if (res.status == 500 || res.status == 404) {
        createErrorPage(res)
      }
      if (!res.ok) {
        throw await res.text()
      }
      return res.text()
    })
    .then(() => {
      const loginData = {
        identifier: formData.nickname,
        password: formData.password,
      }
      fetch("/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(loginData),
      }).then((res) => {
        if (res.status != 200 && res.status != 401 && res.status != 201) {
          createErrorPage(res)
        }
        window.location.reload()
      })
    })
    .catch((err) => {
      errorToast(err)
    })
}

function calculateAge(birthDate) {
  const today = new Date()
  const dob = new Date(birthDate)
  let age = today.getFullYear() - dob.getFullYear()
  const m = today.getMonth() - dob.getMonth()
  if (m < 0 || (m === 0 && today.getDate() < dob.getDate())) {
    age--
  }
  return age
}
