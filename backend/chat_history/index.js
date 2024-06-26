import { shared } from '@appblocks/node-sdk'

const handler = async (event) => {
  const { req, res } = event

  const { prisma, healthCheck, getBody, sendResponse, validateBody, authenticateUser } = await shared.getShared()

  try {
    // health check
    if (healthCheck(req, res)) return
    const reqBody = await getBody(req)

    await validateBody(reqBody, 'chatHistorySchema')

    const userInfo = await authenticateUser(req)

    if (userInfo.error) {
      sendResponse(res, 400, { success: false, msg: userInfo.error })
      return
    }

    const { page = 1, limit = 10, chat_id } = reqBody
    const offset = (page - 1) * limit

    const chatHistoryCount = await prisma.$queryRaw`SELECT COUNT(*) as total FROM (
      SELECT 
      *
      FROM individual_chat_details  as icd
      WHERE icd."chatId" =${BigInt(chat_id)} 
  ) as subquery;
  `

    const chatHistory = await prisma.$queryRaw`
    select * from individual_chat_details as icd
    WHERE icd."chatId" =${BigInt(chat_id)} 
    order by icd."createdAt" desc
    limit ${limit}
    offset ${offset}
    `

    let result = {
      chat_history: chatHistory,
      count: chatHistoryCount[0].total,
    }

    sendResponse(res, 200, { success: true, msg: `chat history retrived successfully`, data: result })
  } catch (error) {
    console.error('Error retriving chat history:', error)
  }
}

export default handler
