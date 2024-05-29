export const isEmpty = (val) => {
  return (
    val == null ||
    (typeof val === "object" && Object.keys(val).length === 0) ||
    (typeof val === "string" && val.trim().length === 0)
  );
};

export const isJson = (val) => {
  try {
    JSON.parse(val);
  } catch (e) {
    return false;
  }
  return true;
};

export const getUrlParams = (windowLocSearch) => {
  return new Proxy(new URLSearchParams(windowLocSearch), {
    get: (searchParams, prop) => searchParams.get(prop),
  });

}
