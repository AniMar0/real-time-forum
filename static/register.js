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

  if (Age < 13 || Age > 120) {
    errorToast("Your age is not accepted. Must be between 13 and 120 years old.");
    return
  }

  if (document.getElementById("nickname").value == "") {
    errorToast("Nickname is required");
    return
  }
  if (document.getElementById("firstName").value == "") {
    errorToast("First name is required");
    return
  }
  if (document.getElementById("lastName").value == "") {
    errorToast("Last name is required");
    return
  }
  if (document.getElementById("email").value == "") {
    errorToast("Email is required");
    return
  }
  if (document.getElementById("password").value == "" || (document.getElementById("password").value).length < 8) {
    errorToast("Password must be at least 8 characters long");
    return
  }
  if (document.getElementById("gender").value == "") {
    errorToast("Gender is required");
    return
  }

  const formData = {
    nickname: document.getElementById("nickname").value,
    first_name: document.getElementById("firstName").value,
    last_name: document.getElementById("lastName").value,
    email: document.getElementById("email").value,
    password: document.getElementById("password").value,
    age: Age,
    gender: document.getElementById("gender").value
  };

  fetch("/register", {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify(formData)
  })
    .then(async res => {
      if (res.status == 500 || res.status == 404) {
        ErrorPage(res)
      }
      if (!res.ok) {
        throw await res.text();
      }
      return res.text();
    })
    .then(() => {
      const loginData = {
        identifier: formData.nickname,
        password: formData.password
      };
      fetch("/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify(loginData)
      }).then((res) => {
        if (res.status != 200 && res.status != 401 && res.status != 201) {
          ErrorPage(res)
        }
        window.location.reload();
      })
    })
    .catch(err => {
      errorToast(err);
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