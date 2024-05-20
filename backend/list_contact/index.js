import { shared } from '@appblocks/node-sdk'

const handler = async (event) => {
  const { req, res } = event

  const { prisma, healthCheck, getBody, sendResponse, validateBody, authenticateUser } = await shared.getShared()

  try {
    // health check
    if (healthCheck(req, res)) return
    const reqBody = await getBody(req)

    await validateBody(reqBody, 'paginationAndSearchSchema')

    const userInfo = await authenticateUser(req)

    if (userInfo.error) {
      sendResponse(res, 400, { success: false, msg: userInfo.error })
      return
    }

    const { page = 1, limit = 10, search = '', project_id } = reqBody
    const skip = (page - 1) * limit
    let searchValue = `%${search}%`

    const contactsCount = await prisma.$queryRaw`SELECT COUNT(*) as total FROM (
      SELECT 
       c.name  
      FROM contacts as c
      WHERE c."projectId" =${project_id}
      AND c.name ILIKE ${searchValue}
  ) as subquery;
  `

    const contactsInfo = await prisma.$queryRaw`
      SELECT 
      *
      FROM contacts as c
      WHERE c."projectId" =${project_id}
      AND c.name ILIKE ${searchValue}
      LIMIT ${limit} 
      OFFSET ${skip};`

    let result = {
      data: contactsInfo,
      count: contactsCount[0].total,
    }

    sendResponse(res, 200, { success: true, msg: `contacts retrived successfully`, data: result })
  } catch (error) {
    console.error('Error retriving contacts:', error)
  }
}

export default handler
