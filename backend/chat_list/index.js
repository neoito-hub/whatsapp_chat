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

    const { page = 1, limit = 10, search = '', project_id } = reqBody
    const skip = (page - 1) * limit
    let searchValue = `%${search}%`

    const chatsCount = await prisma.$queryRaw`SELECT COUNT(*) as total FROM (
      SELECT 
      * 
      FROM chats as c
      WHERE c."projectId" =${project_id}
      AND c."chatName" ILIKE ${searchValue}
  ) as subquery;
  `

    const chatsInfo = await prisma.$queryRaw`
      SELECT 
      *
      FROM chats as c
      WHERE c."projectId" =${project_id}
      AND c."chatName" ILIKE ${searchValue}
      LIMIT ${limit} 
      OFFSET ${skip};`

    let result = {
      chats: chatsInfo,
      count: chatsCount[0].total,
    }

    sendResponse(res, 200, { success: true, msg: `chats retrived successfully`, data: result })
  } catch (error) {
    console.error('Error retriving chats:', error)
  }
}

export default handler
