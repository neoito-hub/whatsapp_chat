import ApiEndpoints from "../config/apiEndPoints.js";

const submitAppnamePermissionForm = () => {
  
  const form = document.querySelector("#authAppnamePermission");
  if (!form) return;

  const fields = [];
    const urlencode = new URL(ApiEndpoints.allowAuthPermission)
    urlencode.search = window.location.search;
    console.log(ApiEndpoints.allowAuthPermission);
    document.getElementById("authAppnamePermission").action = urlencode.toString();
    document.getElementById("authAppnamePermission").method = "post"; 
    document.getElementById("authAppnamePermission").submit()
}

window.submitAppnamePermissionForm = submitAppnamePermissionForm;