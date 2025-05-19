import { renderLoginForm } from './login.js';

export function renderRegisterForm() {
  const app = document.getElementById("app");

  app.innerHTML = `
    <div class="form-container">
      <h2>Register</h2>
      <form id="registerForm">
        <label for="nickname">Nickname:</label>
        <input type="text" id="nickname" name="nickname" required />

        <label for="age">Age:</label>
        <input type="number" id="age" name="age" required />

        <label for="gender">Gender:</label>
        <select id="gender" name="gender" required>
          <option value="">Select</option>
          <option value="Male">Male</option>
          <option value="Female">Female</option>
          <option value="Other">Other</option>
        </select>

        <label for="firstName">First Name:</label>
        <input type="text" id="firstName" name="firstName" required />

        <label for="lastName">Last Name:</label>
        <input type="text" id="lastName" name="lastName" required />

        <label for="email">Email:</label>
        <input type="email" id="email" name="email" required />

        <label for="password">Password:</label>
        <input type="password" id="password" name="password" required />

        <button type="submit">Register</button>
        <p>Already have an account? <a href="#" id="goToLogin">Login here</a></p>
      </form>
    </div>
  `;

  // Attach submit handler
  document.getElementById("registerForm").addEventListener("submit", handleRegister);

  // Link back to login form
  document.getElementById("goToLogin").addEventListener("click", (e) => {
    e.preventDefault();
    renderLoginForm();
  });
}

// Placeholder for registration logic
export function handleRegister(event) {
  document.getElementById("register-form").addEventListener("submit", function (e) {
    e.preventDefault();

    const data = {
      Nickname: document.getElementById("nickname").value,
      Age: parseInt(document.getElementById("age").value),
      Gender: document.getElementById("gender").value,
      FirstName: document.getElementById("first-name").value,
      LastName: document.getElementById("last-name").value,
      Email: document.getElementById("email").value,
      Password: document.getElementById("password").value
    };

    fetch("/register", {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify(data)
    })
      .then(res => {
        if (res.ok) {
          window.location.href = "/login";
        } else {
          alert("Something went wrong");
        }
      })
      .catch(err => console.error("Error:", err));
  });

}