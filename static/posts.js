export async function loadPosts() {
  const response = await fetch('/posts');
  const posts = await response.json();

  const postsList = document.getElementById('postsList');
  postsList.innerHTML = '';
  if (posts != null) {
    posts.forEach(post => {
      const div = document.createElement('div');
      div.classList.add('post');
      div.innerHTML = `
      <h3>${post.title}</h3>
      <p>${post.content}</p>
      <small>Category: ${post.category} | By: ${post.author} | At: ${new Date(post.created_at).toLocaleString()}</small>
    `;
      postsList.appendChild(div);
    });
  }
}


