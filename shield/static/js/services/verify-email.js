import ApiEndpoints from "../config/apiEndPoints.js";
import { postData } from "../fetch/index.js";
import { printError } from "../utils/errorHandler.js";
import { getUrlParams } from "../utils/helper.js";

var timeLeft = 32;
var elem = document.getElementById('Timer');

var timerId = setInterval(countdown, 1000);

function countdown() {
  if (timeLeft == 0) {
    clearTimeout(timerId);
    elem.className='w-full text-center text-sm text-purple mt-3 cursor-pointer hover:underline'
    elem.innerHTML ='Resend OTP'
    setTimeout(function() {
      document
      .getElementsByClassName("success-msg-wrapper")[0]
      .classList.add("hidden");
  }, 1000);
    document.getElementById("Timer").addEventListener("click", function(){
      if(timeLeft===0){
        clearTimeout(timerId);
      elem.className='text-gray-light'
       timeLeft = 32;
       timerId=setInterval(countdown, 1000);
       const params = getUrlParams(window.location.search);
    const userId=params.user_id
    const userIdParsed={user_id:userId}

    document.getElementsByClassName("error-msg-wrapper")[0].classList.add("hidden");

    postData(ApiEndpoints.resendOTP,userIdParsed).then((data) => {
      if (!data) return;
    });
      }
    })
  } else {
    elem.innerHTML ='Resend OTP(0:'+timeLeft+'s)'
    timeLeft--;
  }
}

const verifyEmail = (e) => {
  e.preventDefault();
  const form = document.querySelector("#formVerifyEmail");
  if (!form) return false;

  const input = document.querySelectorAll("input.otp-input");
  let verification_code = "";

  input.forEach((item) => {
    verification_code = `${verification_code}${item.value}`;
  });

  if (!verification_code || verification_code.length != 6) {
    printError(null, "Please enter the correct otp");
    return false;
  }

  const params = getUrlParams(window.location.search);

  const user_id = params.user_id;

  var input1 = document.createElement('input');
  input1.type = 'hidden';
  input1.name = 'user_id';
  input1.value = user_id;
  form.appendChild(input1);

  var input2 = document.createElement('input');
  input2.type = 'hidden';
  input2.name = 'verification_code';
  input2.value = verification_code;
  form.appendChild(input2);

    const urlencode = new URL(ApiEndpoints.verifyEmail)
    urlencode.search = window.location.search;

    form.action = urlencode.toString();
    form.method = "post";
    form.submit();
};

let digitValidate = function (ele) {
  ele.value = ele.value.replace(/[^0-9]/g, "");
};

let tabChange = function (val) {
  let ele = document.querySelectorAll("input.otp-input");
  if (ele[val - 1].value.length >= 3) {
    ele[1].focus();
  } else if (ele[val - 1].value == "") {
    ele[0].focus();
  }
};

window.digitValidate = digitValidate;
window.tabChange = tabChange;
window.verifyEmail = verifyEmail;
