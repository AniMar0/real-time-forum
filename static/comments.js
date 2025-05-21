export async function loadComments(postId) {
    try {
      const res = await fetch(`/comments?postId=${postId}`);
      const comments = await res.json();
  
      const container = document.getElementById(`comments-${postId}`);
      container.innerHTML = '';
  
      comments.forEach(comment => {
        const el = document.createElement('p');
        el.textContent = `${comment.username}: ${comment.content}`;
        container.appendChild(el);
      });
    } catch (err) {
      console.error(err);
    }
  }
  
  export function setupCommentSubmission() {
    document.querySelectorAll('.commentForm').forEach(form => {
      form.addEventListener('submit', async (e) => {
        e.preventDefault();
  
        const postId = form.dataset.postId;
        const content = form.comment.value;
  
        const res = await fetch('/addComment', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({ postId, content })
        });
  
        if (res.ok) {
          form.reset();
          loadComments(postId);
        } else {
          alert('Failed to add comment');
        }
      });
    });
  }
  