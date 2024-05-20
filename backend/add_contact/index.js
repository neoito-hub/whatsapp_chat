import { shared } from '@appblocks/node-sdk'
import { nanoid } from 'nanoid'

const handler = async (event) => {
  const { req, res } = event

  const { prisma, healthCheck, getBody, sendResponse, validateBody, validate_phone_number, authenticateUser } =
    await shared.getShared()

  try {
    // health check
    if (healthCheck(req, res)) return
    const reqBody = await getBody(req)
    
    await validateBody(reqBody, 'addContactSchema')

    const userInfo = await authenticateUser(req)

    if (userInfo.error) {
      sendResponse(res, 400, { success: false, msg: userInfo.error })
      return;
    }


    const getValidatedPhoneNumber = await validate_phone_number(reqBody.phoneNumber, reqBody.countryCode)

    // Check if the phone number already exists in the database
    const existingContact = await prisma.contacts.findUnique({
      where: {
        phoneNumber: reqBody.phoneNumber,
      },
    })

    if (existingContact) {
      sendResponse(res, 400, { success: false, msg: `Phone number already exists` })
    }

    const newContact = await prisma.contacts.create({
      data: { id: nanoid(), ...reqBody },
    })

    const parameters = [
      {
        contactId: newContact.id,
        name: 'name',
        value: getValidatedPhoneNumber.phone_number_with_country_code,
        status: 'active',
      },
      {
        contactId: newContact.id,
        name: 'phone',
        value: getValidatedPhoneNumber.phone_number_with_country_code,
        status: 'active',
      },
    ]

    await prisma.candidate_custom_parameters.createMany({
      data: parameters,
    })

    sendResponse(res, 200, { success: true, msg: `contact added successfully`, data: newContact })
  } catch (error) {
    console.error('Error adding contact:', error)
  }
}

export default handler
