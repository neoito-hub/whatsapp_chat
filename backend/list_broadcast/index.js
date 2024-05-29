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

    const { page = 1, limit = 10, search = '' } = reqBody

    const skip = (page - 1) * limit

    let searchValue = `%${search}%`

    const broadcastsCount = await prisma.$queryRaw`SELECT COUNT(*) as total FROM (
      SELECT 
          b.id,
          b."templateMessageId",
          b.name,
          b."totalNumberOfReceipients",
          b.broadcast_template_params,
          b.header_params,
          b.footer_params,
          b.body_params,
          b."broadcastTime",
          b."createdAt",
          b."updatedAt",
          b."project_id",
          json_agg(json_build_object('id', br.id, 'user_id', br."userId", 'name', c.name, 'received_status', br."recievedStatus")) AS Recipients
      FROM broadcasts as b
      LEFT JOIN broadcast_recipients as br ON b.id = br."broadcastId"
      LEFT JOIN contacts as c ON br."userId" = c.id 
      WHERE b."project_id" =${reqBody.project_id}
      AND b.name ILIKE ${searchValue}
      GROUP BY 
      b.id,
          b."templateMessageId",
          b.name,
          b."totalNumberOfReceipients",
          b.broadcast_template_params,
          b.header_params,
          b.footer_params,
          b.body_params,
          b."broadcastTime",
          b."createdAt",
          b."updatedAt",
          b."project_id"
    ) as subquery;
    `

    const broadcasts = await prisma.$queryRaw`
    SELECT 
          b.id,
          b."templateMessageId",
          b.name,
          b."totalNumberOfReceipients",
          b.broadcast_template_params,
          b.header_params,
          b.footer_params,
          b.body_params,
          b."broadcastTime",
          b."createdAt",
          b."updatedAt",
          b."project_id",
          json_agg(json_build_object('id', br.id, 'user_id', br."userId", 'name', c.name, 'received_status', br."recievedStatus")) AS Recipients
      FROM broadcasts as b
      LEFT JOIN broadcast_recipients as br ON b.id = br."broadcastId"
      LEFT JOIN contacts as c ON br."userId" = c.id 
      WHERE b."project_id" =${reqBody.project_id}
      AND b.name ILIKE ${searchValue}
      GROUP BY 
      b.id,
          b."templateMessageId",
          b.name,
          b."totalNumberOfReceipients",
          b.broadcast_template_params,
          b.header_params,
          b.footer_params,
          b.body_params,
          b."broadcastTime",
          b."createdAt",
          b."updatedAt",
          b."project_id"
    LIMIT ${limit} 
    OFFSET ${skip};`

    let result = {
      broadcasts: broadcasts,
      count: broadcastsCount[0].total,
    }

    sendResponse(res, 200, { success: true, msg: `Broadcasts retrived successfully`, data: result })
  } catch (error) {
    console.error('Error retriving Broadcasts:', error)
  }
}

export default handler
