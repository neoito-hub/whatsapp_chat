import ApiEndpoints from "../config/apiEndPoints.js";
import FormSubmitHandler from "../utils/formSubmitHandler.js";

const emailInput = document.getElementById('emaillogin');

emailInput.addEventListener('input', () => {
	emailInput.value = emailInput.value.replace(/^\s*(.*)$/, (wholeString, captureGroup) => captureGroup);
});
const passwordInput = document.getElementById('password');

passwordInput.addEventListener('input', (event) => {
	passwordInput.value = passwordInput.value.replace(/^\s*(.*)$/, (wholeString, captureGroup) => captureGroup);
});

class Login {
  constructor() {
    this.submitLoginForm();
  }
  submitLoginForm() {
    const form = document.querySelector("#loginForm");
    if (!form) return;

    const fields = ["emaillogin", "password"];

    const formSubmitInstance = new FormSubmitHandler(form, fields);
    
    formSubmitInstance.validateOnSubmit().then((loginData) => {
      if (!loginData) return this.submitLoginForm();
      const urlencode = new URL(ApiEndpoints.login)
      urlencode.search = window.location.search;
      console.log(ApiEndpoints.login);
      document.getElementById("loginForm").action = urlencode.toString();
      document.getElementById("loginForm").method = "post"; 
      document.getElementById("loginForm").submit()
    });
  }
}
new Login();
