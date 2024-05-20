import { shared } from '@appblocks/node-sdk'

const handler = async (event) => {
  const { req, res } = event

  const { prisma, healthCheck, getBody, sendResponse, validateBody, authenticateUser } = await shared.getShared()

  try {
    // health check
    if (healthCheck(req, res)) return
    const reqBody = await getBody(req)

    await validateBody(reqBody, 'deleteSchema')

    const userInfo = await authenticateUser(req)

    if (userInfo.error) {
      sendResponse(res, 400, { success: false, msg: userInfo.error })
      return;
    }

    let contactId = reqBody.id

    // Check if the user exists in the database
    const existingContact = await prisma.contacts.findUnique({
      where: {
        id: contactId,
      },
    })

    if (!existingContact) {
      sendResponse(res, 400, { success: false, msg: `Contact not exists` })
    }

    const deletedContact = await prisma.contacts.delete({
      where: {
        id: contactId,
      },
    })

    sendResponse(res, 200, { success: true, msg: `contact deleted successfully`, data: deletedContact })
  } catch (error) {
    console.error('Error deleting contact:', error)
  }
}

export default handler
