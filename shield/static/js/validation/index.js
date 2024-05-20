import { REGEX } from "../utils/constants.js";
import { printError, removeAllErrors } from "../utils/errorHandler.js";

export default class Validation {
  constructor() {}

  validateAndGetData(fields) {
    const self = this;
    const data = {};
    let isError = false;
    fields.forEach((field) => {
      if (isError) return;
      const input = document.querySelector(`#${field}`);
    
      if (self.validateFields(input) == false) isError = true;
      else if (
        field === "passwordRe" &&
        data.password &&
        input.value !== data.password
      ) {
        printError("passwordRe", "Password does not match");
        isError = true;
      } else {
        data[field] = input.value.trim();
      }
    });

    return { isError, data };
  }

  validateFields(field) {
    removeAllErrors();

    const fieldValue = field.value.trim();
    const fieldId = field.id;
    const fieldFrom = field.form.id


    if (fieldValue === "") {
      console.log(fieldId,"fieldValue")
      printError(fieldId, `Please enter your ${fieldId==='emaillogin'?'Username/Email':fieldId==='passwordRe'?'Re-enter Password':fieldId.charAt(0).toUpperCase()+ fieldId.slice(1)}`);

      return false;
    }


    if (fieldId == "acceptTerms" && field.checked === false) {
      printError("acceptTerms", "Please accept the Terms and Conditions");
      return false;
    }
    if (fieldId === "email" && REGEX.email.test(fieldValue) === false) {
      printError(fieldId, "Please enter a Valid Email Address");
      return false;
    }
    if (fieldId === "password" && fieldFrom != "loginForm" && REGEX.password.test(fieldValue) === false) {
      printError(
        fieldId,
        `Password must have minimum eight characters, at least one Uppercase letter, one Lowercase letter, one Number and one Special Character`
      );
      return false;
    }
    return true;
  }
}
