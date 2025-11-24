import { setupCommentSubmission, toggleComments } from "./comments.js"
import { createErrorPage } from "./error.js"
import { successToast, errorToast } from "./toast.js"

export function createPostsSection() {
  const section = document.createElement("section")
  section.id = "postsSection"

  const h2 = document.createElement("h2")
  h2.textContent = "Posts"

  const postsContainer = document.createElement("div")
  postsContainer.id = "postsContainer"

  // Create post form
  const createPostForm = document.createElement("form")
  createPostForm.id = "createPostForm"

  const titleInput = document.createElement("input")
  titleInput.name = "title"
  titleInput.placeholder = "Post Title"
  titleInput.required = true

  const contentTextarea = document.createElement("textarea")
  contentTextarea.name = "content"
  contentTextarea.placeholder = "Write your post here..."
  contentTextarea.required = true

  const categorySelect = document.createElement("select")
  categorySelect.name = "category"

  const generalOption = document.createElement("option")
  generalOption.value = "General"
  generalOption.textContent = "General"

  const questionsOption = document.createElement("option")
  questionsOption.value = "Questions"
  questionsOption.textContent = "Questions"

  const newsOption = document.createElement("option")
  newsOption.value = "News"
  newsOption.textContent = "News"

  categorySelect.appendChild(generalOption)
  categorySelect.appendChild(questionsOption)
  categorySelect.appendChild(newsOption)

  const submitBtn = document.createElement("button")
  submitBtn.type = "submit"
  submitBtn.textContent = "Post"

  createPostForm.appendChild(titleInput)
  createPostForm.appendChild(contentTextarea)
  createPostForm.appendChild(categorySelect)
  createPostForm.appendChild(submitBtn)

  // Add form submit handler
  createPostForm.addEventListener("submit", async (e) => {
    e.preventDefault()

    const form = e.target
    const postData = {
      title: form.title.value,
      content: form.content.value,
      category: form.category.value,
    }

    const response = await fetch("/createPost", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(postData),
      credentials: "include",
    })

    if (response.ok) {
      successToast("Post created successfully!")
      form.reset()
      loadPosts()
    } else {
      console.log("Failed to create post:", response)
      errorToast((await response.text()) || "Failed to create post.")
    }
  })

  // Posts list
  const postsList = document.createElement("div")
  postsList.id = "postsList"

  postsContainer.appendChild(createPostForm)
  postsContainer.appendChild(postsList)

  section.appendChild(h2)
  section.appendChild(postsContainer)

  return section
}

export async function loadPosts() {
  const response = await fetch("/posts")

  if (response.status != 200 && response.status != 401 && response.status != 201) {
    createErrorPage(response)
  }

  const posts = await response.json()

  const postsList = document.getElementById("postsList")
  postsList.innerHTML = ""

  posts.forEach((post) => {
    const div = document.createElement("div")
    div.classList.add("post")

    const h3 = document.createElement("h3")
    h3.textContent = post.title

    const p = document.createElement("p")
    p.textContent = post.content

    const small = document.createElement("small")
    small.textContent = `Category: ${post.category} | By: ${post.author} | At: ${new Date(post.created_at).toLocaleString()}`

    const postActions = document.createElement("div")
    postActions.classList.add("post-actions")

    const toggleBtn = document.createElement("button")
    toggleBtn.classList.add("toggle-comments-btn")
    toggleBtn.setAttribute("data-post-id", post.id)
    toggleBtn.textContent = "Show Comments"

    postActions.appendChild(toggleBtn)

    // Comments section
    const commentsSection = document.createElement("div")
    commentsSection.id = `comments-section-${post.id}`
    commentsSection.classList.add("comments-section", "hidden")

    const h4 = document.createElement("h4")
    h4.textContent = "Comments"

    const commentsContainer = document.createElement("div")
    commentsContainer.id = `comments-${post.id}`
    commentsContainer.classList.add("comments-container")

    const loadingP = document.createElement("p")
    loadingP.textContent = "Loading comments..."
    commentsContainer.appendChild(loadingP)

    // Comment form
    const commentForm = document.createElement("form")
    commentForm.id = `comment-form-${post.id}`
    commentForm.classList.add("comment-form")

    const commentTextarea = document.createElement("textarea")
    commentTextarea.classList.add("comment-input")
    commentTextarea.placeholder = "Write a comment..."
    commentTextarea.required = true

    const commentSubmitBtn = document.createElement("button")
    commentSubmitBtn.type = "submit"
    commentSubmitBtn.textContent = "Post Comment"

    commentForm.appendChild(commentTextarea)
    commentForm.appendChild(commentSubmitBtn)

    commentsSection.appendChild(h4)
    commentsSection.appendChild(commentsContainer)
    commentsSection.appendChild(commentForm)

    div.appendChild(h3)
    div.appendChild(p)
    div.appendChild(small)
    div.appendChild(postActions)
    div.appendChild(commentsSection)

    postsList.appendChild(div)

    toggleBtn.addEventListener("click", () => {
      const postId = toggleBtn.getAttribute("data-post-id")
      toggleComments(postId)
      if (toggleBtn.textContent.trim() === "Show Comments") {
        toggleBtn.textContent = "Hide Comments"
      } else {
        toggleBtn.textContent = "Show Comments"
      }
    })
    setupCommentSubmission(post.id)
  })
}
