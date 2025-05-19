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
  event.preventDefault();

  const user = {
    nickname: document.getElementById("nickname").value,
    age: document.getElementById("age").value,
    gender: document.getElementById("gender").value,
    firstName: document.getElementById("firstName").value,
    lastName: document.getElementById("lastName").value,
    email: document.getElementById("email").value,
    password: document.getElementById("password").value,
  };

  console.log("Registering user:", user);

  // server send data 
}