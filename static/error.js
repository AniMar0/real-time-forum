export function ErrorPage(data) {
    const root = document.getElementById('root')
    if (!root) return

    const errorPage = document.createElement('main')
    errorPage.id = 'errorPage'

    const heading = document.createElement('h1')
    heading.textContent = String(data?.status ?? 500)

    const message = document.createElement('p')
    message.textContent = data?.statusText || 'Something went wrong.'

    const backButton = document.createElement('button')
    backButton.type = 'button'
    backButton.textContent = 'Back to home'
    backButton.addEventListener('click', () => {
        window.location.href = '/'
    })

    errorPage.append(heading, message, backButton)
    root.replaceChildren(errorPage)
}
