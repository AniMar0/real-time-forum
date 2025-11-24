import { showSection } from "./app.js"
import { createErrorPage } from "./error.js"
import { errorToast } from "./toast.js"

export async function loadComments(postId) {
  try {
    const response = await fetch(`/comments?post_id=${postId}`)
    if (response.status != 200 && response.status != 401 && response.status != 201) {
      createErrorPage(response)
    }
    if (!response.ok) {
      throw new Error("Failed to load comments")
    }
    const comments = await response.json()
    displayComments(postId, comments)
  } catch (error) {
    errorToast("Failed to load comments")
  }
}

function displayComments(postId, comments) {
  const commentsContainer = document.getElementById(`comments-${postId}`)
  if (!commentsContainer) return
  commentsContainer.innerHTML = ""
  if (!comments || comments.length === 0) {
    const noCommentsP = document.createElement("p")
    noCommentsP.classList.add("no-comments")
    noCommentsP.textContent = "No comments yet. Be the first to comment!"
    commentsContainer.appendChild(noCommentsP)
    return
  }
  comments.forEach((comment) => {
    const commentElement = document.createElement("div")
    commentElement.classList.add("comment")

    const commentHeader = document.createElement("div")
    commentHeader.classList.add("comment-header")

    const authorSpan = document.createElement("span")
    authorSpan.classList.add("comment-author")
    authorSpan.textContent = comment.author

    const dateSpan = document.createElement("span")
    dateSpan.classList.add("comment-date")
    dateSpan.textContent = new Date(comment.created_at).toLocaleString()

    commentHeader.appendChild(authorSpan)
    commentHeader.appendChild(dateSpan)

    const contentDiv = document.createElement("div")
    contentDiv.classList.add("comment-content")
    contentDiv.textContent = comment.content

    commentElement.appendChild(commentHeader)
    commentElement.appendChild(contentDiv)

    commentsContainer.appendChild(commentElement)
  })
}

export function setupCommentSubmission(postId) {
  const form = document.getElementById(`comment-form-${postId}`)
  if (!form) return
  form.addEventListener("submit", async (e) => {
    e.preventDefault()
    const commentContent = form.querySelector(".comment-input").value.trim()
    if (!commentContent) return
    try {
      const response = await fetch("/createComment", {
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

      if (response.status != 200 && response.status != 401 && response.status != 201) {
        createErrorPage(response)
      }

      if (!response.ok) {
        throw new Error("Failed to submit comment")
      }

      form.querySelector(".comment-input").value = ""
      loadComments(postId)
    } catch (error) {
      showSection("login")
    }
  })
}

export function toggleComments(postId) {
  const commentsSection = document.getElementById(`comments-section-${postId}`)
  if (commentsSection.classList.contains("hidden")) {
    commentsSection.classList.remove("hidden")
    loadComments(postId)
  } else {
    commentsSection.classList.add("hidden")
  }
}
