import ApiEndpoints from "../config/apiEndPoints.js";
import { postData } from "../fetch/index.js";
import FormSubmitHandler from "../utils/formSubmitHandler.js";

var elem = document.getElementById('close');
elem.addEventListener('click', () => {
  document
  .getElementsByClassName("error-msg-wrapper")[0]
  .classList.add("hidden");
});

class PasswordRecovery {
  constructor() {
    this.submitPasswordRecoveryForm();
  }
  
  submitPasswordRecoveryForm() {
    const form = document.querySelector("#passwordRecoveryForm");
    if (!form) return;

    const fields = ["email"];
    const formSubmitInstance = new FormSubmitHandler(form, fields);
    formSubmitInstance.validateOnSubmit().then((passwordRecoveryData) => {
    if (!passwordRecoveryData) return this.submitPasswordRecoveryForm();

      postData(ApiEndpoints.passwordRecovery, passwordRecoveryData).then((data) => {
        if (!data) return this.submitPasswordRecoveryForm();
        setTimeout(function() {
          document
          .getElementsByClassName("success-msg-wrapper")[0]
          .classList.add("hidden");
      }, 10000);
      document.getElementById("passwordRecoveryForm").reset();
      setTimeout(function() {
        window.location.reload()}, 10000);
      });
    });
  }
  
}
new PasswordRecovery();
