import { handleRegister } from './register.js';
import { handleLogin } from './login.js';
import { logout } from './logout.js';
import { logged } from './app.js';
import { successToast, errorToast } from './toast.js';
import { loadPosts } from './posts.js';

export function clearRoot() {
    const root = document.getElementById("root");
    root.innerHTML = "";
    return root;
}

export function renderLoginPage() {
    const root = clearRoot();

    const section = document.createElement("section");
    section.id = "loginSection";

    section.innerHTML = `
        <h2>Login</h2>
        <form id="loginForm">
          <input id="identifier" placeholder="Email or Nickname" required />
          <input id="loginPassword" placeholder="Password" type="password" required />
          <div id="loginError" style="color: red; display: none; margin-bottom: 10px;"></div>
          <button type="submit">Login</button>
        </form>
        <p>Don't have an account? <button id="showRegister" type="button">Register</button></p>
    `;

    root.appendChild(section);

    // Attach event listeners AFTER elements are created
    document.getElementById('loginForm').addEventListener('submit', handleLogin);
    document.getElementById('showRegister').addEventListener('click', renderRegisterPage);
}

export function renderRegisterPage() {
    const root = clearRoot();

    const section = document.createElement("section");
    section.id = "registerSection";

    section.innerHTML = `
        <h2>Register</h2>
        <form id="registerForm">
          <input id="nickname" placeholder="Nickname" required />
          <input id="firstName" placeholder="First Name" required />
          <input id="lastName" placeholder="Last Name" required />
          <input id="email" placeholder="Email" type="email" required />
          <input id="password" placeholder="Password" type="password" required />
          <input type="date" id="dateOfBirth" required>
          <select id="gender" required>
            <option value="">Select Gender</option>
            <option value="male">Male</option>
            <option value="female">Female</option>
          </select>
          <button type="submit">Register</button>
        </form>
        <p>Already have an account? <button id="showLogin" type="button">Login</button></p>
    `;

    root.appendChild(section);

    // Attach event listeners AFTER elements are created
    document.getElementById('registerForm').addEventListener('submit', handleRegister);
    document.getElementById('showLogin').addEventListener('click', renderLoginPage);
}

export function renderLoggedPage(username) {
    const root = clearRoot();

    // HEADER
    const header = document.createElement("header");
    header.innerHTML = `
      <h1>My Forum</h1>
      <nav>
        <span id="usernameDisplay">${username}</span>
        <button id="logoutBtn">Logout</button>
      </nav>
    `;
    root.appendChild(header);

    const main = document.createElement("main");
    main.className = "flex-container";

    main.innerHTML = `
      <div id="userList"></div>

      <div class="main-content">

        <div id="chatWindow" class="chat-box hidden">
          <div class="chat-header">
            <strong>Chat with: <span id="chatWithName"></span></strong>
            <i id="closeChatBtn" class="fa-solid fa-xmark" style="cursor:pointer;"></i>
          </div>

          <div id="chatLoader" class="hidden" style="text-align:center; padding:5px;">
            <i class="fa fa-spinner fa-spin"></i> Loading more...
          </div>

          <div id="chatMessages"></div>
          <div id="typingIndicator" class="hidden"></div>

          <input id="messageInput" type="text" placeholder="Type a message..." />
          <button id="sendBtn">Send</button>
        </div>

        <section id="postsSection">
          <h2>Posts</h2>
          <div id="postsContainer">
            <form id="createPostForm">
              <input name="title" placeholder="Post Title" required />
              <textarea name="content" placeholder="Write your post here..." required></textarea>
              <select name="category">
                <option value="General">General</option>
                <option value="Questions">Questions</option>
                <option value="News">News</option>
              </select>
              <button type="submit">Post</button>
            </form>

            <div id="postsList"></div>
          </div>
        </section>

      </div>
    `;

    root.appendChild(main);

    // Attach event listeners AFTER elements are created
    document.getElementById('logoutBtn').addEventListener('click', (e) => {
        logout(e);
        document.getElementById('usernameDisplay').textContent = logged(false);
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
            successToast('Post created successfully!');
            form.reset();
            loadPosts();
        } else {
            console.log('Failed to create post:', response);
            errorToast(await response.text() || 'Failed to create post.');
        }
    });
}