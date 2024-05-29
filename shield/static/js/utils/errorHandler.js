import { isEmpty } from "./helper.js";

export const openErrorWrapper = () => {
  document
    .getElementsByClassName("error-msg-wrapper")[0]
    .classList.remove("hidden");
};

export const printError = (elemId, errMsg) => {
  document.getElementById("error-msg").innerHTML = errMsg;
  if (elemId) document.getElementById(elemId).classList.add("has-error");
  openErrorWrapper();
};

export const removeAllErrors = () => {
  const errorClass = document.getElementsByClassName("has-error");
  for (let errorClasses of errorClass) {
    errorClasses.classList.remove("has-error");
  }
};

export const clearErrorWrapper = () => {
  document.getElementById("error-msg").innerHTML = "";
  document
    .getElementsByClassName("error-msg-wrapper")[0]
    .classList.add("hidden");
};

export const resErrorHandler = (errorData) => {
  const { err, message, data } = errorData;
  if (data) {
    if (!data.includes("ERROR")) {
      return printError(null, data);
    }

    const key = data
      .replace("_key", "")
      ?.match(/"([^"]+)"/)[1]
      ?.split("_")
      ?.slice(1)
      ?.join(" ");
    const errMsg = data.toLowerCase().includes("duplicate")
      ? "is duplicate"
      : "is wrong";

    return printError(
      null,
      key && errMsg ? `${key} ${errMsg}` : data.errMessage
    );
  }
  if (isEmpty(err)) return printError(null, message);

  Object.entries(err).forEach(([field, subErrorMsg]) => {
    printError(field, subErrorMsg);
  });
};

export const unhandledErrorHandler = (errors) => {
  printError(null, "Something went wrong");
  console.log({ errors });
};
