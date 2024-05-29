import shield from './shieldUtils.js'
/**
 * Function to format and send response
 * @param {*} res
 * @param {*} code
 * @param {*} data
 * @param {*} type
 */
// const sendResponse = (res, code, data, type = 'application/json') => {
//   const headers = {
//     'Access-Control-Allow-Origin': '*',
//     'Access-Control-Allow-Methods': 'GET, POST, OPTIONS, PUT, DELETE',
//     'Content-Type': type,
//   }

//   res.writeHead(code, headers)
//   res.write(JSON.stringify(data))
//   res.end()
// }

const sendResponse = (res, code, data, type = 'application/json') => {
  const headers = {
    'Access-Control-Allow-Origin': '*',
    'Access-Control-Allow-Methods': 'GET, POST, OPTIONS, PUT, DELETE',
    'Content-Type': type,
  }

  // Custom JSON replacer function to handle BigInt values
  const bigIntReplacer = (key, value) => {
    if (typeof value === 'bigint') {
      return value.toString() // Convert BigInt to string
    }
    return value
  }

  const responseData = JSON.stringify(data, bigIntReplacer)
  headers['Content-Length'] = Buffer.byteLength(responseData, 'utf-8')

  res.writeHead(code, headers)
  res.write(responseData, 'utf-8', (err) => {
    if (err) {
      console.error('Error while sending response:', err)
    }
    res.end()
  })
}

/**
 * Function to extract the body from the request
 * @param {*} req
 * @returns
 */
const getBody = async (req) => {
  const bodyBuffer = []
  for await (const chunk of req) {
    bodyBuffer.push(chunk)
  }
  const data = Buffer.concat(bodyBuffer).toString()
  return JSON.parse(data || '{}')
}

const healthCheck = (req, res) => {
  if (req.params.health === 'health') {
    sendResponse(res, 200, { success: true, message: 'Health check success' })
    return true
  }
  return false
}

const validate_phone_number = async (phoneNumber, countryCode) => {
  const sanitized_country_code = String(countryCode).replace(/[^0-9]/g, '')
  const sanitized_phone_number = String(phoneNumber).replace(/[^0-9]/g, '')
  const phone_number_with_country_code = `${sanitized_country_code}${sanitized_phone_number}`
  const phone_number_without_country_code = sanitized_phone_number
  return { phone_number_with_country_code, phone_number_without_country_code }
}

const authenticateUser = async (req) => {
  try {
    // Get user details using shield
    const userDetails = await shield.getUser(req)
    return { id: userDetails.user_id }
  } catch (e) {
    return { error: 'Authentication Error' }
  }
}

// const authenticateUser = async (req) => {
//   const user = await getUser(req);
//   console.log("user is", user);
//   const authHeader = req.headers.get("authorization");
//   const token = authHeader && authHeader.split(" ")[1];

//   try {
//     if (token == null) throw new Error();
//     const data = jwt.verify(token, process.env.BB_AUTH_SECRET_KEY.toString());
//     return data;
//   } catch (e) {
//     console.log("error is", e);
//     const error = new Error("An error occurred.");
//     error.errorCode = 401;
//     throw error;
//   }
// };

export default { healthCheck, getBody, sendResponse, validate_phone_number, authenticateUser }
