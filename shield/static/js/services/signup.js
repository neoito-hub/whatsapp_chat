import ApiEndpoints from "../config/apiEndPoints.js";
import FormSubmitHandler from "../utils/formSubmitHandler.js";

// var modal = document.getElementById("myModal");

// // Get the button that opens the modal
// var btn = document.getElementById("myBtn");
// var elem = document.getElementById('or');

// btn.addEventListener("click", function(){
//   elem.classList.add("hidden")
// })

// // Get the <span> element that closes the modal
// var span = document.getElementsByClassName("close")[0];

// // When the user clicks on the button, open the modal
// btn.onclick = function() {
//   modal.style.display = "block";
// }

// // When the user clicks on <span> (x), close the modal
// span.onclick = function() {
//   modal.style.display = "none";
//   elem.classList.remove("hidden")

// }

// // When the user clicks anywhere outside of the modal, close it
// window.onclick = function(event) {
//   if (event.target == modal) {
//     modal.style.display = "none";

//   }
// }

const userNameInput = document.getElementById('username');

userNameInput.addEventListener('input', () => {
	userNameInput.value = userNameInput.value.replace(/^\s*(.*)$/, (wholeString, captureGroup) => captureGroup);
});
const emailInput = document.getElementById('email');

emailInput.addEventListener('input', () => {
	emailInput.value = emailInput.value.replace(/^\s*(.*)$/, (wholeString, captureGroup) => captureGroup);
});
const passwordInput = document.getElementById('password');

passwordInput.addEventListener('input', (event) => {
	passwordInput.value = passwordInput.value.replace(/^\s*(.*)$/, (wholeString, captureGroup) => captureGroup);
});

class Signup {
  constructor() {
    this.submitSignupForm();
  }
  


  submitSignupForm() {
   
    const form = document.querySelector("#signupForm");
    if (!form) return;

    const fields = ["email", "username", "password", "acceptTerms"];
    const formSubmitInstance = new FormSubmitHandler(form, fields);

    formSubmitInstance.validateOnSubmit().then((signupData) => {
      if (!signupData) return this.submitSignupForm();

      const urlencode = new URL(ApiEndpoints.signup)
      urlencode.search = window.location.search;
      console.log(ApiEndpoints.signup);
      document.getElementById("signupForm").action = urlencode.toString();
      document.getElementById("signupForm").method = "post"; 
      document.getElementById("signupForm").submit()
    });
    return false;
  }
}

new Signup();
