import { shared } from '@appblocks/node-sdk'

const handler = async (event) => {
  const { req, res } = event

  const { prisma, healthCheck, sendResponse, authenticateUser } = await shared.getShared()

  try {
    // health check
    if (healthCheck(req, res)) return

    const userInfo = await authenticateUser(req)

    if (userInfo.error) {
      sendResponse(res, 400, { success: false, msg: userInfo.error })
      return
    }

    const languages = await prisma.template_languages.findMany()

    sendResponse(res, 200, { success: true, msg: `languages retrived successfully`, data: languages })
  } catch (error) {
    console.error('Error retriving languages:', error)
  }
}

export default handler
