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

  // Attach submit event handler
  document.getElementById("loginForm").addEventListener("submit", handleLogin);

  // Link to registration page
  document.getElementById("goToRegister").addEventListener("click", (e) => {
    e.preventDefault();
    renderRegisterForm();
  });
}

// Placeholder for login logic
export function handleLogin(event) {
  event.preventDefault();

  const identifier = document.getElementById("identifier").value;
  const password = document.getElementById("password").value;

  console.log("Login with", identifier, password);

  // server conection
}


