import {
  unhandledErrorHandler,
  resErrorHandler,
} from "../utils/errorHandler.js";
import { isJson } from "../utils/helper.js";
import { resSuccessHandler } from "../utils/successHandler.js";

const commonFetchOptions = {
  mode: "cors", // no-cors, *cors, same-origin
  // cache: "no-cache", // default, no-cache, reload, force-cache, only-if-cached
  // credentials: "same-origin", // include, *same-origin, omit
  // redirect: "follow", // manual, *follow, error
  // referrerPolicy: "no-referrer", // no-referrer, *no-referrer-when-downgrade, origin, origin-when-cross-origin, same-origin, strict-origin, strict-origin-when-cross-origin, unsafe-url
};

// Example POST method implementation:
export const postData = async (url = "", data = {}) => {
  // Default options are marked with *
  try {
    const urlencode = new URL(url);
    urlencode.search = window.location.search;
    const response = await fetch(urlencode, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Access-Control-Allow-Origin": "*",
      },
      body: JSON.stringify(data), // body data type must match "Content-Type" header
      ...commonFetchOptions,
    });


    if (response.redirected || response.status > 300 && response.status < 400) {
      window.location.href = response.url;
      return null;
    }

    const resData = await response.json(); // parses JSON response into native JavaScript objects

    // Handle html response
    if (resData.html) {
      document.write(resData.html);
      return null;
    }

    if (resData.success)
    {
      console.log({ resData });
      resSuccessHandler(resData)
      return resData;
    } 
else{
    // Handle Error response
    resErrorHandler(resData);
    return false;

}
  
  } catch (err) {
    //Handle any other error
    unhandledErrorHandler(err);
    return false;
  }
};

// Example GET method implementation:
export const getData = async (url = "", params = {}) => {
  // Default options are marked with *
  try {
    const urlencode = new URL(url);
    // urlencode.search = new URLSearchParams(params);
    urlencode.search = window.location.search;

    const response = await fetch(urlencode, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        "Access-Control-Allow-Origin": "*",
      },
      ...commonFetchOptions,
    });

    //alert(JSON.stringify({response}))

    // if (!isJson(response)){
    //   alert("in if !isJson")
    //   window.location.href = response.url
    //   //document.write(await response.text());
    //   return null;
    // }

    const resData = await response.json(); // parses JSON response into native JavaScript objects
    console.log({ resData });

    // Handle html response
    if (resData.html) {
      document.write(resData.html);
      return null;
    }

    if (resData.success) return resData;

    // Handle Error response
    resErrorHandler(resData);
    return false;
  } catch (err) {
    //Handle any other error
    unhandledErrorHandler(err);
    return false;
  }
};
