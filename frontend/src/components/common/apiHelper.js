/* eslint-disable no-unused-expressions */
/* eslint-disable consistent-return */
import axios from 'axios'
import { shield } from '@appblocks/js-sdk'
// import { toast } from "react-toastify";

const apiHelper = async ({
  baseUrl,
  subUrl,
  value = null,
  apiType = 'post',
  // showSuccessMessage = false,
  // spaceId,
}) => {
  const token = shield.tokenStore.getToken()
  try {
    const { data } = await axios({
      method: apiType,
      url: `${baseUrl}${subUrl}`,
      data: value && value,
      headers: token && {
        Authorization: `Bearer ${token}`,
        // space_id: spaceId,
      },
    })
    // showSuccessMessage && toast.success(data?.msg);
    return data?.data
  } catch (err) {
    console.log(err?.response?.data?.msg)
    // toast.error(err?.response?.data?.msg);
    // if (err.response.status === 401) shield.logout();
  }
}

export default apiHelper
