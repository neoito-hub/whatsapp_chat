import { shared } from '@appblocks/node-sdk'

const handler = async (event) => {
  const { req, res } = event

  const { prisma, healthCheck, getBody, sendResponse, validateBody, authenticateUser } = await shared.getShared()

  try {
    // health check
    if (healthCheck(req, res)) return
    const reqBody = await getBody(req)

    await validateBody(reqBody, 'getTemplateDetails')

    const userInfo = await authenticateUser(req)

    if (userInfo.error) {
      sendResponse(res, 400, { success: false, msg: userInfo.error })
      return
    }

    const templateInfo = await prisma.$queryRaw`
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
    (
    SELECT json_agg(json_build_object(
      'id', tmcp.id, 
      'paramName', tmcp."paramName",
      'paramValue', tmcp."paramValue",
      'paramOrder', tmcp."paramOrder",
      'paramType', tmcp."paramType"
    ))
    FROM template_message_custom_parameters tmcp
    WHERE tmcp."templateMessageId" = t.id AND tmcp.id IS NOT NULL
  ) AS template_message_custom_parameters,
	  t."projectId",
    (
    SELECT json_agg(json_build_object(
      'id', tmb.id, 'buttonArray', tmb."buttonArray"
    ))
    FROM template_message_buttons tmb
    WHERE tmb."templateMessageId" = t.id AND tmb.id IS NOT NULL
  ) AS template_message_buttons
FROM templates as t
LEFT JOIN template_message_buttons as tmb ON t.id = tmb."templateMessageId"
LEFT JOIN template_message_custom_parameters as tmcp ON t.id = tmcp."templateMessageId"
WHERE t."projectId" =${reqBody.project_id}
AND t.id =${reqBody.template_id}
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
    t."projectId" `

    sendResponse(res, 200, { success: true, msg: `template retrived successfully`, data: templateInfo })
  } catch (error) {
    console.error('Error retriving templates:', error)
  }
}

export default handler
