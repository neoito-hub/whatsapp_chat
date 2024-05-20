import { shared } from '@appblocks/node-sdk'

const handler = async (event) => {
  const { req, res } = event

  const { prisma, healthCheck, getBody, sendResponse, validateBody, authenticateUser } = await shared.getShared()

  try {
    // health check
    if (healthCheck(req, res)) return
    const reqBody = await getBody(req)

    await validateBody(reqBody, 'chatlistSchema')

    const userInfo = await authenticateUser(req)

    if (userInfo.error) {
      sendResponse(res, 400, { success: false, msg: userInfo.error })
      return
    }

    const { page = 1, limit = 10, project_id, search } = reqBody

    const skip = (page - 1) * limit
    const searchValue = `%${search}%`

    const whereCondition = {
      projectId: project_id,
      chatName: {
        contains: searchValue,
        mode: 'insensitive',
      },
    }

    const chats = await prisma.chats.findMany({
      where: whereCondition,
      skip,
      take: limit,
    })

    sendResponse(res, 200, { success: true, msg: `chats retrived successfully`, data: chats })
  } catch (error) {
    console.error('Error retriving chats:', error)
  }
}

export default handler
