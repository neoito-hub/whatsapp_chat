import { shared } from '@appblocks/node-sdk'

const handler = async (event) => {
  const { req, res } = event

  const { prisma, healthCheck, getBody, sendResponse, validateBody, authenticateUser } = await shared.getShared()

  try {
    // health check
    if (healthCheck(req, res)) return
    const reqBody = await getBody(req)

    await validateBody(reqBody, 'editContactSchema')

    const userInfo = await authenticateUser(req)

    if (userInfo.error) {
      sendResponse(res, 400, { success: false, msg: userInfo.error })
      return;
    }

    let contactId = reqBody.id
    let name = reqBody?.name
    let countryCode = reqBody?.countryCode
    let phoneNumber = reqBody?.phoneNumber
    let email = reqBody?.email
    let address = reqBody?.address

    // Check if the user exists in the database
    const existingContact = await prisma.contacts.findUnique({
      where: {
        id: contactId,
      },
    })

    if (!existingContact) {
      sendResponse(res, 400, { success: false, msg: `Contact not exists` })
    }

    const updatedContact = await prisma.contacts.update({
      where: {
        id: contactId,
      },
      data: {
        name,
        countryCode,
        phoneNumber,
        email,
        address,
      },
    })

    sendResponse(res, 200, { success: true, msg: `contact updated successfully`, data: updatedContact })
  } catch (error) {
    console.error('Error adding contact:', error)
  }
}

export default handler
