import { handleRegister } from './regester.js';
import { handleLogin } from './login.js';
import { loadPosts } from './posts.js';
import { logout } from './logout.js';

export function showSection(sectionId) {
  document.querySelectorAll('section').forEach(section => {
    section.classList.add('hidden');
  });
  document.getElementById(sectionId).classList.remove('hidden');
}

document.getElementById('showLogin').addEventListener('click', () => {
  showSection('loginSection');
});

document.getElementById('showRegister').addEventListener('click', () => {
  showSection('registerSection');
});

document.getElementById('logoutBtn').addEventListener('click', (e) => {
  logout(e);
  document.getElementById('showLogin').classList.remove('hidden');
  document.getElementById('showRegister').classList.remove('hidden');
  document.getElementById('logoutBtn').classList.add('hidden');
});

document.getElementById('registerForm').addEventListener('submit', async function (e) {
  handleRegister(e);
});

document.getElementById('loginForm').addEventListener('submit', async function (e) {
  handleLogin(e);
  document.getElementById('showLogin').classList.add('hidden');
  document.getElementById('showRegister').classList.add('hidden');
  document.getElementById('logoutBtn').classList.remove('hidden');
});

document.getElementById('createPostForm').addEventListener('submit', async function (e) {
  e.preventDefault();

  const form = e.target;
  const postData = {
    title: form.title.value,
    content: form.content.value,
    category: form.category.value
  };

  const response = await fetch('/createPost', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(postData),
    credentials: 'include'
  });

  if (response.ok) {
    alert('Post created!');
    form.reset();
    loadPosts();
  } else {
    alert('Failed to create post');
  }
});

document.addEventListener('DOMContentLoaded', function () {
  loadPosts();

});