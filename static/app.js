import { startChatFeature } from './chat.js';
import { loadPosts } from './posts.js';
import { ErrorPage } from './error.js';
import { renderLoggedPage, renderLoginPage } from './dom.js';

const checkLoggedIn = () => {
  fetch('/logged', {
    method: 'POST',
    credentials: 'include'
  })
    .then(res => {
      if (res.status != 200 && res.status != 401 && res.status != 405 && res.status != 201) {
        ErrorPage(res)
      }
      if (!res.ok) throw new Error('Not logged in')
      return res.json()
    })
    .then(data => {
      logged(true, data.username)
      renderLoggedPage(data.username)
      startChatFeature(data.username)
      loadPosts()
    })
    .catch(() => {
      logged(false)
      renderLoginPage()
    })
}

// Global timer to check every minute if user is still logged in
setInterval(() => {
  checkLoggedIn();
}, 60000); // 60000ms = 1 minute

document.addEventListener('DOMContentLoaded', function () {
  if (window.location.pathname != "/") {
    ErrorPage({ status: 404, statusText: "Page not found" })
  }
  checkLoggedIn();
});

window.addEventListener('storage', function (event) {
  if (event.key === 'logout') {
    window.location.reload()
  }
});

export function showSection(sectionId) {
  document.querySelectorAll('section').forEach(section => {
    section.classList.add('hidden');
  });
  const targetSection = document.getElementById(sectionId);
  if (targetSection) {
    targetSection.classList.remove('hidden');
  }
}

export function logged(bool, user) {
  const usernameDisplay = document.getElementById('usernameDisplay');
  const showLogin = document.getElementById('showLogin');
  const showRegister = document.getElementById('showRegister');
  const logoutBtn = document.getElementById('logoutBtn');
  const createPostForm = document.getElementById('createPostForm');
  const userList = document.getElementById('userList');

  if (bool) {
    if (usernameDisplay) {
      usernameDisplay.textContent = user;
      usernameDisplay.classList.remove('hidden');
    }
    if (showLogin) showLogin.classList.add('hidden');
    if (showRegister) showRegister.classList.add('hidden');
    if (logoutBtn) logoutBtn.classList.remove('hidden');
    if (createPostForm) createPostForm.classList.remove('hidden');
    if (userList) userList.classList.remove('hidden');
  } else {
    if (usernameDisplay) {
      usernameDisplay.textContent = "";
      usernameDisplay.classList.add('hidden');
    }
    if (showLogin) showLogin.classList.remove('hidden');
    if (showRegister) showRegister.classList.remove('hidden');
    if (logoutBtn) logoutBtn.classList.add('hidden');
    if (createPostForm) createPostForm.classList.add('hidden');
    if (userList) userList.classList.add('hidden');
  }
}