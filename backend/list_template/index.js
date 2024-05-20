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

    const templatesCount = await prisma.$queryRaw`SELECT COUNT(*) as total FROM (
      SELECT 
          t.id,
          t.name,
          t."categoryId",
          t."categoryName",
          t."languageId",
          t."languageName",
          t."templateType",
          t."templateBody",
          t."templateHeader",
          t."templateFooter",
          t."templateUuid",
          t.status,
          t."buttonsType",
          t.type,
          t."createdAt",
          t."updatedAt",
          t."projectId",
          json_agg(json_build_object('id', tmcp.id, 'paramName', tmcp."paramName",'paramValue',tmcp."paramValue",
           'paramOrder',tmcp."paramOrder",'paramType',tmcp."paramType")) AS template_message_custom_params,
            t."projectId",
          json_agg(json_build_object('id', tmb.id, 'buttonArray', tmb."buttonArray")) AS template_message_buttons    
      FROM templates as t
      LEFT JOIN template_message_buttons as tmb ON t.id = tmb."templateMessageId"
      LEFT JOIN template_message_custom_parameters as tmcp ON t.id = tmcp."templateMessageId"
      WHERE t."projectId" =${reqBody.project_id}
      AND t.name ILIKE ${searchValue}
      GROUP BY 
          t.id,
          t.name,
          t."categoryId",
          t."categoryName",
          t."languageId",
          t."languageName",
          t."templateType",
          t."templateBody",
          t."templateHeader",
          t."templateFooter",
          t."templateUuid",
          t.status,
          t."buttonsType",
          t.type,
          t."createdAt",
          t."updatedAt",
          t."projectId"
  ) as subquery;
  `

    const templatesInfo = await prisma.$queryRaw`
    SELECT 
    t.id,
    t.name,
    t."categoryId",
    t."categoryName",
    t."languageId",
    t."languageName",
    t."templateType",
    t."templateBody",
    t."templateHeader",
    t."templateFooter",
    t."templateUuid",
    t.status,
    t."buttonsType",
    t.type,
    t."createdAt",
    t."updatedAt",
    t."projectId",
    json_agg(json_build_object('id', tmcp.id, 'paramName', tmcp."paramName",'paramValue',tmcp."paramValue",
     'paramOrder',tmcp."paramOrder",'paramType',tmcp."paramType")) AS template_message_custom_params,
	  t."projectId",
    json_agg(json_build_object('id', tmb.id, 'buttonArray', tmb."buttonArray")) AS template_message_buttons	 
FROM templates as t
LEFT JOIN template_message_buttons as tmb ON t.id = tmb."templateMessageId"
LEFT JOIN template_message_custom_parameters as tmcp ON t.id = tmcp."templateMessageId"
WHERE t."projectId" =${reqBody.project_id}
AND t.name ILIKE ${searchValue}
GROUP BY 
    t.id,
    t.name,
    t."categoryId",
    t."categoryName",
    t."languageId",
    t."languageName",
    t."templateType",
    t."templateBody",
    t."templateHeader",
    t."templateFooter",
    t."templateUuid",
    t.status,
    t."buttonsType",
    t.type,
    t."createdAt",
    t."updatedAt",
    t."projectId" 
    LIMIT ${limit} 
    OFFSET ${skip};`

    let result = {
      data: templatesInfo,
      count: templatesCount[0].total,
    }

    sendResponse(res, 200, { success: true, msg: `templates retrived successfully`, data: result })
  } catch (error) {
    console.error('Error retriving templates:', error)
  }
}

export default handler
