import { showSection, logged } from './app.js';
import { startChatFeature } from './chat.js';
import { ErrorPage } from './error.js';
import { loadPosts } from './posts.js';



export function handleLogin(event) {
  event.preventDefault()
  const loginError = document.getElementById("loginError")
  loginError.style.display = "none"

  const formData = {
    identifier: document.getElementById("identifier").value,
    password: document.getElementById("loginPassword").value
  };

  fetch("/login", {
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
      return res.json();
    })
    .then(data => {
      startChatFeature(data.username);
      loadPosts();
      showSection('postsSection');
      logged(true, data.username);
    })
    .catch(err => {
      loginError.textContent = "Incorrect email/nickname or password."
      loginError.style.display = "block"
      logged(false)
      console.error(err)
    })
}
