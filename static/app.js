import {renderLoginForm} from './login.js';

export function showPosts(event) {
    if (event) {
        document.getElementById("Postes").hidden = false;
    }else{
        document.getElementById("Postes").hidden = true;
    }
    
}

export function showApps(event) {
    if (event) {
        document.getElementById("app").hidden = false;
    }else{
        document.getElementById("app").hidden = true;
    }
}

window.addEventListener("DOMContentLoaded", () => {
  renderLoginForm();
});