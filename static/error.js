export function createErrorPage(res) {
    const app = document.getElementById("app")
    app.innerHTML = "" // Clear everything
  
    const errPage = document.createElement("div")
    errPage.id = "err-page"
  
    const errCode = document.createElement("h1")
    errCode.id = "err-code"
    errCode.textContent = res.status
  
    const errMsj = document.createElement("h2")
    errMsj.id = "err-msj"
    errMsj.textContent = res.statusText
  
    const backBtn = document.createElement("button")
    backBtn.id = "back"
    backBtn.textContent = "back"
    backBtn.addEventListener("click", () => {
      window.location.reload()
    })
  
    errPage.appendChild(errCode)
    errPage.appendChild(errMsj)
    errPage.appendChild(backBtn)
  
    app.appendChild(errPage)
  }
  