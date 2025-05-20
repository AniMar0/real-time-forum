import { showSection } from './app.js';

export function handleRegister(event) {
  event.preventDefault();

  const formData = {
    nickname: document.getElementById("nickname").value,
    first_name: document.getElementById("firstName").value,
    last_name: document.getElementById("lastName").value,
    email: document.getElementById("email").value,
    password: document.getElementById("password").value,
    age: parseInt(document.getElementById("age").value),
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
      if (!res.ok) {
        throw new Error("Registration failed");
      }
      return res.text();
    })
    .then(data => {
      alert("Registration successful");
      showSection('loginSection');
    })
    .catch(err => {
      alert("Error: " + err.message);
      console.error(err);
    });
}
