import Validation from "../validation/index.js";
import { clearErrorWrapper } from "./errorHandler.js";

export default class FormSubmit extends Validation {
  constructor(form, fields) {
    super();
    this.form = form;
    this.fields = fields;
  }

  validateOnSubmit() {
    return new Promise((resolve) => {
      this.form.addEventListener("submit", (e) => {
        e.preventDefault();
        e.stopPropagation();

        const { isError, data } = this.validateAndGetData(this.fields);
        if (isError) return resolve(false);
        clearErrorWrapper();
        resolve(data);
      });
    });
  }
}


  // Test get all form values from form
  // const form = document.querySelector('form')
  // Object.values(form).reduce((obj,field) => { obj[field.name] = field.value; return obj }, {})
