import {renderRegisterForm} from './regester.js';
export function renderLoginForm() {
  const app = document.getElementById("app");

  app.innerHTML = `
    <div class="form-container">
      <h2>Login</h2>
      <form id="loginForm">
        <label for="identifier">Email or Nickname:</label>
        <input type="text" id="identifier" name="identifier" required />

        <label for="password">Password:</label>
        <input type="password" id="password" name="password" required />

        <button type="submit">Login</button>
        <p>Don't have an account? <a href="#" id="goToRegister">Register here</a></p>
      </form>
    </div>
  `;

  document.getElementById("loginForm").addEventListener("submit", handleLogin);

  document.getElementById("goToRegister").addEventListener("click", (e) => {
    e.preventDefault();
    renderRegisterForm();
  });
}

export function handleLogin(event) {
  event.preventDefault();


  const form = event.target;
  const formData = {
    identifier: form.identifier.value,
    password: form.password.value,
  };
  
  fetch("/login", {
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
      renderLoginForm();
    })
    .catch(err => {
      alert("Error: " + err.message);
      console.error(err);
    });
}


