export function ErrorPage(data) {
    const MainPage = document.getElementById('my-content')
    const ErrPage = document.getElementById('err-page')
    const ErrorCode = document.getElementById('err-code')
    const ErrorMsj = document.getElementById('err-msj')
    const BackBtn = document.getElementById('back')

    MainPage.classList.add('hidden');
    ErrPage.classList.remove('hidden');

    ErrorCode.innerText = data.status
    ErrorMsj.innerText = data.statusText

    BackBtn.addEventListener('click', () => {
        MainPage.classList.remove('hidden');
        ErrPage.classList.add('hidden');
        history.pushState({}, '', '/');
    })
}