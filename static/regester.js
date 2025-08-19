import { showSection } from './app.js';

export function handleRegister(event) {
  event.preventDefault();

  const Age = parseInt(document.getElementById("age").value);

  if (Age < 13 || Age > 120) {
    alert("you age is not accepted")
    return
  }

  if (document.getElementById("nickname").value == "") {
    alert("nickname empty")
    return
  }
  if (document.getElementById("firstName").value == "") {
    alert("firstName empty")
    return
  }
  if (document.getElementById("lastName").value == "") {
    alert("lastName empty")
    return
  }
  if (document.getElementById("email").value == "") {
    alert("email empty")
    return
  }
  if (document.getElementById("password").value == "" || (document.getElementById("password").value).length < 8) {
    alert("email empty or <8")
    return
  }
  if (document.getElementById("gender").value == "") {
    alert("gender empty")
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
    .then(res => {
      if (res.status != 200 && res.status != 401) {
        ErrorPage(res)
      }
      if (!res.ok) {
        throw new Error("Registration failed");
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
        if (res.status != 200 && res.status != 401) {
          ErrorPage(res)
        }
        window.location.reload();
      })
    })
    .catch(err => {
      console.error(err);
    });
}
