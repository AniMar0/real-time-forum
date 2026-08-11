import { handleRegister } from './register.js';
import { handleLogin } from './login.js';
import { logout } from './logout.js';
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
    section.className = "auth-page";

    section.innerHTML = `
        <div class="auth-intro">
          <span class="eyebrow">MY FORUM</span>
          <h1>Ideas move faster when people talk.</h1>
          <p>Join focused conversations, share what you know, and keep the signal high.</p>
          <div class="intro-note"><span class="intro-dot"></span> A calmer space for useful discussions</div>
        </div>
        <div class="auth-card">
          <div class="auth-heading">
            <span class="brand-mark">F</span>
            <div><h2>Welcome back</h2><p>Sign in to continue the conversation.</p></div>
          </div>
          <form id="loginForm">
            <label for="identifier">Email or nickname</label>
            <input id="identifier" placeholder="you@example.com" autocomplete="username" required />
            <label for="loginPassword">Password</label>
            <input id="loginPassword" placeholder="Enter your password" type="password" autocomplete="current-password" required />
            <div id="loginError"></div>
            <button type="submit">Sign in <span aria-hidden="true">→</span></button>
          </form>
          <p class="auth-switch">New to the forum? <button id="showRegister" type="button">Create an account</button></p>
        </div>
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
    section.className = "auth-page";

    section.innerHTML = `
        <div class="auth-intro">
          <span class="eyebrow">MY FORUM</span>
          <h1>Bring your perspective to the room.</h1>
          <p>Create an account and find conversations worth returning to.</p>
          <div class="intro-note"><span class="intro-dot"></span> Your next good idea starts here</div>
        </div>
        <div class="auth-card auth-card-wide">
          <div class="auth-heading">
            <span class="brand-mark">F</span>
            <div><h2>Create your account</h2><p>A few details and you are ready to join.</p></div>
          </div>
          <form id="registerForm" class="register-grid">
            <div><label for="nickname">Nickname</label><input id="nickname" placeholder="your_nickname" autocomplete="username" required /></div>
            <div><label for="email">Email</label><input id="email" placeholder="you@example.com" type="email" autocomplete="email" required /></div>
            <div><label for="firstName">First name</label><input id="firstName" placeholder="First name" autocomplete="given-name" required /></div>
            <div><label for="lastName">Last name</label><input id="lastName" placeholder="Last name" autocomplete="family-name" required /></div>
            <div><label for="password">Password</label><input id="password" placeholder="At least 8 characters" type="password" autocomplete="new-password" required /></div>
            <div><label for="dateOfBirth">Date of birth</label><input type="date" id="dateOfBirth" required /></div>
            <div><label for="gender">Gender</label><select id="gender" required><option value="">Select gender</option><option value="male">Male</option><option value="female">Female</option></select></div>
            <button type="submit">Create account <span aria-hidden="true">→</span></button>
          </form>
          <p class="auth-switch">Already have an account? <button id="showLogin" type="button">Sign in</button></p>
        </div>
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
      <div class="brand-lockup"><span class="brand-mark">F</span><div><h1>My Forum</h1><small>Make room for better ideas.</small></div></div>
      <nav>
        <span class="user-greeting">Signed in as <strong id="usernameDisplay"></strong></span>
        <button id="logoutBtn">Log out</button>
      </nav>
    `;
    header.querySelector('#usernameDisplay').textContent = username;
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
          <div class="section-heading"><div><span class="eyebrow">COMMUNITY FEED</span><h2>Latest discussions</h2><p>Share something useful with the community.</p></div></div>
          <div id="postsContainer">
            <form id="createPostForm">
              <div class="composer-title"><span class="composer-avatar">${String(username).charAt(0).toUpperCase()}</span><input name="title" placeholder="Give your post a clear title" required /></div>
              <textarea name="content" placeholder="What would you like to discuss?" required></textarea>
              <select name="category">
                <option value="General">General</option>
                <option value="Questions">Questions</option>
                <option value="News">News</option>
              </select>
              <button type="submit">Publish post <span aria-hidden="true">→</span></button>
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
    });

    document.getElementById('createPostForm').addEventListener('submit', async function (e) {
        e.preventDefault();

        const form = e.target;
        const title = form.elements.namedItem('title').value.trim();
        const content = form.elements.namedItem('content').value.trim();
        const category = form.elements.namedItem('category').value.trim();
        if (!title || !content || !category) {
            errorToast('Title, content, and category are required.');
            return;
        }

        const postData = {
            title,
            content,
            category
        };

        try {
            const response = await fetch('/createPost', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(postData),
                credentials: 'include'
            });

            if (!response.ok) {
                errorToast(await response.text() || 'Failed to create post.');
                return;
            }

            successToast('Post created successfully!');
            form.reset();
            loadPosts();
        } catch (err) {
            errorToast('Failed to create post. Please try again.');
        }
    });
}
