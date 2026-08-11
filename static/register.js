import { showSection } from './app.js';
import { ErrorPage } from './error.js';
import { errorToast } from './toast.js';

export function handleRegister(event) {
  event.preventDefault();

  const dob = document.getElementById("dateOfBirth").value;
  if (!dob) {
    errorToast("Your age is not accepted");
    return;
  }
  const Age = calculateAge(dob);

  const nickname = document.getElementById("nickname").value.trim();
  const firstName = document.getElementById("firstName").value.trim();
  const lastName = document.getElementById("lastName").value.trim();
  const email = document.getElementById("email").value.trim();
  const password = document.getElementById("password").value;
  const gender = document.getElementById("gender").value;

  if (Age < 13 || Age > 120) {
    errorToast("Your age is not accepted. Must be between 13 and 120 years old.");
    return
  }

  if (!/^[a-zA-Z0-9_]{3,20}$/.test(nickname)) {
    errorToast("Nickname must be 3-20 characters using letters, numbers, or underscores");
    return
  }
  if (firstName.length < 1 || firstName.length > 50) {
    errorToast("First name must be 1-50 characters");
    return
  }
  if (lastName.length < 1 || lastName.length > 50) {
    errorToast("Last name must be 1-50 characters");
    return
  }
  if (!/^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/.test(email)) {
    errorToast("Enter a valid email address");
    return
  }
  if (password.length < 8 || !/[A-Z]/.test(password) || !/[a-z]/.test(password) || !/[0-9]/.test(password)) {
    errorToast("Password must be at least 8 characters with uppercase, lowercase, and number");
    return
  }
  if (gender !== "male" && gender !== "female") {
    errorToast("Gender is required");
    return
  }

  const formData = {
    nickname,
    first_name: firstName,
    last_name: lastName,
    email,
    password,
    age: Age,
    gender
  };

  fetch("/register", {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify(formData)
    })
    .then(async res => {
      if (!res.ok) {
        if (res.status >= 500) ErrorPage(res)
        throw new Error(await res.text() || "Registration failed");
      }
      return res.text();
    })
    .then(() => {
      const loginData = {
        identifier: formData.nickname,
        password: formData.password
      };
      return fetch("/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify(loginData)
      }).then(async (res) => {
        if (!res.ok) {
          if (res.status >= 500) ErrorPage(res)
          throw new Error(await res.text() || "Automatic login failed");
        }
        window.location.reload();
      })
    })
    .catch(err => {
      errorToast(err instanceof Error ? err.message : String(err));
    });
}

function calculateAge(birthDate) {
  const today = new Date();
  const dob = new Date(birthDate);
  let age = today.getFullYear() - dob.getFullYear();
  const m = today.getMonth() - dob.getMonth();
  if (m < 0 || (m === 0 && today.getDate() < dob.getDate())) {
    age--;
  }
  return age;
}
