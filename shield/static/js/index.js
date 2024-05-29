function typeChanger(inputID, _this) {
  var elm = document.getElementById(inputID);
  if (elm.type === "password") {
    elm.type = "text";
    _this.classList.add("active-icon");
  } else {
    elm.type = "password";
    _this.classList.remove("active-icon");
  }
}
function handleClosebtn() {
  document
    .getElementsByClassName("error-msg-wrapper")[0]
    .classList.add("hidden");
}

function handleOnClick() {
  let displayMessage = '';
  const error = document.getElementById("errorMessage");
    error.innerText = displayMessage;
}

window.typeChanger = typeChanger;
window.handleClosebtn = handleClosebtn;
window.handleOnClick = handleOnClick;

