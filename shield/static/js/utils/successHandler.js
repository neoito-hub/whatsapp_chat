import { isEmpty } from "./helper.js";

export const openSucessWrapper = () => {
    document
      .getElementsByClassName("success-msg-wrapper")[0]
      .classList.remove("hidden");
  };

export const printSuccess = (elemId, successMsg) => {
    console.log(successMsg,"innn1")
  document.getElementById("success-msg").innerHTML = successMsg;
  if (elemId) document.getElementById(elemId).classList.add("success");
  openSucessWrapper();
};

export const resSuccessHandler = (resData) => {
  const { success, message, data } = resData;
  console.log(resData,"innn1")
  if (data) {
    if (!data.includes("Success")) {
      return printSuccess(null, data);
    }

    const key = data
      .replace("_key", "")
      ?.match(/"([^"]+)"/)[1]
      ?.split("_")
      ?.slice(1)
      ?.join(" ");

    return printSuccess(
      null,
      key && errMsg ? `${key} ${errMsg}` : data.errMessage
    );
  }
  if (isEmpty(success)) return printSuccess(null, message);

  Object.entries(success).forEach(([field, subErrorMsg]) => {
    printSuccess(field, subErrorMsg);
  });
};

