import { ErrorPage } from './error.js';

export function logout(event) {
  event.preventDefault();
  fetch("/logout", {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    }
  })
    .then(res => {
      if (res.status != 200 && res.status != 401 && res.status != 201) {
        ErrorPage(res)
      }
      if (!res.ok) throw new Error("logout failed");
      return res.text();
    })
    .then(() => {
      localStorage.setItem('logout', Date.now());
      window.location.reload()
    })
    .catch(err => {
      alert("Logout failed. Please try again.");
    });
}
