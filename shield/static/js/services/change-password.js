import ApiEndpoints from "../config/apiEndPoints.js";
import { postData } from "../fetch/index.js";
import FormSubmitHandler from "../utils/formSubmitHandler.js";

class ChangePassword {
  constructor() {
    this.submitChangePasswordForm();
  }
  submitChangePasswordForm() {
    const form = document.querySelector("#changePasswordForm");
    if (!form) return;

    const fields = ["password", "passwordRe"];
    
    const formSubmitInstance = new FormSubmitHandler(form, fields);
    formSubmitInstance.validateOnSubmit().then((changePasswordData) => {
      if (!changePasswordData) return this.submitChangePasswordForm();

      const urlencode = new URL(ApiEndpoints.changePassword)
      urlencode.search = window.location.search;
      document.getElementById("changePasswordForm").action = urlencode.toString();
      document.getElementById("changePasswordForm").method = "post"; 
      document.getElementById("changePasswordForm").submit()
    });
  }
}
new ChangePassword();
