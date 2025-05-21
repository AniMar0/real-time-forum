// Function to load comments for a specific post
export async function loadComments(postId) {
    try {
      const response = await fetch(`/comments/${postId}`)
      if (!response.ok) {
        throw new Error("Failed to load comments")
      }
      const comments = await response.json()
      displayComments(postId, comments)
    } catch (error) {
      console.error("Error loading comments:", error)
    }
  }
  
  // Function to display comments for a post
  function displayComments(postId, comments) {
    const commentsContainer = document.getElementById(`comments-${postId}`)
    if (!commentsContainer) return
  
    commentsContainer.innerHTML = ""
  
    if (comments.length === 0) {
      commentsContainer.innerHTML = '<p class="no-comments">No comments yet. Be the first to comment!</p>'
      return
    }
  
    comments.forEach((comment) => {
      const commentElement = document.createElement("div")
      commentElement.classList.add("comment")
      commentElement.innerHTML = `
        <div class="comment-header">
          <span class="comment-author">${comment.author}</span>
          <span class="comment-date">${new Date(comment.created_at).toLocaleString()}</span>
        </div>
        <div class="comment-content">${comment.content}</div>
      `
      commentsContainer.appendChild(commentElement)
    })
  }
  
  // Function to set up comment submission for a post
  export function setupCommentSubmission(postId) {
    const form = document.getElementById(`comment-form-${postId}`)
    if (!form) return
  
    form.addEventListener("submit", async (e) => {
      e.preventDefault()
  
      const commentContent = form.querySelector(".comment-input").value.trim()
      if (!commentContent) return
  
      try {
        const response = await fetch("/comments", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            post_id: postId,
            content: commentContent,
          }),
          credentials: "include",
        })
  
        if (!response.ok) {
          throw new Error("Failed to submit comment")
        }
  
        // Clear the input field
        form.querySelector(".comment-input").value = ""
  
        // Reload comments to show the new one
        loadComments(postId)
      } catch (error) {
        console.error("Error submitting comment:", error)
      }
    })
  }
  
  // Function to toggle comment section visibility
  export function toggleComments(postId) {
    const commentsSection = document.getElementById(`comments-section-${postId}`)
    if (commentsSection.classList.contains("hidden")) {
      commentsSection.classList.remove("hidden")
      loadComments(postId)
    } else {
      commentsSection.classList.add("hidden")
    }
  }
  