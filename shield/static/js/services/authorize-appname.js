import ApiEndpoints from "../config/apiEndPoints.js";
import { getData, postData } from "../fetch/index.js";


// TODO functionality
const allowAuthorization = () => {
    console.log("in func allowAuthorization")
    
    const optionalPermissionList = document
    .getElementById("optionalPermissionList")
    .getElementsByTagName("input");

    const checkedOptions = Array.from(optionalPermissionList).reduce((acc, child) => {
        if (child.checked) acc.push({ id: child.id });
        return acc;
      }, []);

    const urlencode = new URL(ApiEndpoints.allowAppAuthPermissions)
    urlencode.search = window.location.search;
    document.getElementById("authPermission").action = urlencode.toString();
    document.getElementById("authPermission").method = "post"; 
    document.getElementById("authPermission").submit()

};

window.allowAuthorization = allowAuthorization;




const appendPermissionChild = (permissions) => {
    permissions.forEach(({ id, name, description }) => {
      const node = document.createElement("div");
      node.className = "flex mb-3 items-start";
      node.innerHTML = `
      <img
        class="mt-1"
        src="/assets/img/icons/green-tick.svg"
        width="17px"
        height="12px"
        alt=""
      />
      <div class="flex-grow pl-3" id=${id}>
        <div class="text-sm font-semibold text-gray-dark">${name}</div>
        ${
          description
            ? `<div class="text-xs font-normal text-gray-light tracking-tight">
              ${description}
            </div>`
            : ""
        }
      </div>`;
      document.getElementById("permissionList").appendChild(node);
    });
  };
  
  const appendOptionalPermissionChild = (optionalPermissions) => {
    optionalPermissions.forEach(({ id, name }) => {
      const node = document.createElement("label");
      node.className =
        "chk-permission flex text-gray-dark font-semibold text-sm mb-4 cursor-pointer";
      node.innerHTML = `
        <input id=${id} name=${id} checked type="checkbox" class="hidden"/>
        <span class="w-5 h-5 rounded-full flex-shrink-0 border-2 border-green mr-2"></span>${name}
       `;
      document.getElementById("optionalPermissionList").appendChild(node);
    });
  };
  
  // TODO functionality
  const initialize = () => {
    getData(ApiEndpoints.getAllAppPermissionsForAllow).then((resData) => {
        if (!resData) return;
        
        const { mandatory, optional } = resData.data.reduce(
            (acc, data) => {
              if (data.mandatory) acc.mandatory.push(data);
              else acc.optional.push(data);
              return acc;
            },
            { mandatory: [], optional: [] }
          );

         appendPermissionChild(mandatory);
        appendOptionalPermissionChild(optional);

      });

    
  };
  initialize();