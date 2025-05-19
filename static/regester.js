import { renderLoginForm } from './login.js';

export function renderRegisterForm() {
  const app = document.getElementById("app");

  app.innerHTML = `
    <div class="form-container">
      <h2>Register</h2>
      <form id="registerForm" action="/register" method="POST">
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

  document.getElementById("registerForm").addEventListener("submit", handleRegister);

  document.getElementById("goToLogin").addEventListener("click", (e) => {
    e.preventDefault();
    renderLoginForm();
  });
}

export function handleRegister(event) {
  event.preventDefault();

  const form = event.target;
  const formData = {
    nickname: form.nickname.value,
    first_name: form.firstName.value,
    last_name: form.lastName.value,
    email: form.email.value,
    password: form.password.value,
    age: parseInt(form.age.value),
    gender: form.gender.value
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
      renderLoginForm();
    })
    .catch(err => {
      alert("Error: " + err.message);
      console.error(err);
    });
}
