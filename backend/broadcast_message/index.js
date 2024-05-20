import { shared } from '@appblocks/node-sdk'
import { nanoid } from 'nanoid'
import moment from 'moment'
import axios from 'axios'

const handler = async (event) => {
  const { req, res } = event

  const { prisma, healthCheck, getBody, sendResponse, validateBody, validate_phone_number, authenticateUser } =
    await shared.getShared()

  try {
    // health check
    if (healthCheck(req, res)) return
    const reqBody = await getBody(req)

    await validateBody(reqBody, 'newBroadcast')

    const userInfo = await authenticateUser(req)

    if (userInfo.error) {
      sendResponse(res, 400, { success: false, msg: userInfo.error })
      return;
    }

    let actualRecipientIds = []

    for (const userId of reqBody.recipients) {
      actualRecipientIds.push(userId)
    }

    //template info deatils

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
    tmb.id AS buttonId,
    tmb."templateMessageId" AS buttonTemplateMessageId,
    tmb."buttonArray",
    tmcp.id AS paramId,
    tmcp."templateMessageId" AS paramTemplateMessageId,
    tmcp."paramName",
    tmcp."paramValue",
    tmcp."paramOrder",
    tmcp."paramType"
  FROM templates as t
  LEFT JOIN template_message_buttons as tmb ON t.id = tmb."templateMessageId"
  LEFT JOIN template_message_custom_parameters as tmcp ON t.id = tmcp."templateMessageId"
  WHERE t.id = ${reqBody.templateId}
`

    await prisma.$transaction(async (prisma) => {
      const newBroadcast = await prisma.broadcasts.create({
        data: {
          templateMessageId: reqBody.templateId,
          name: reqBody.name,
          broadcastTime: moment.utc(new Date()).toString(),
          successCount: 0,
          failedCount: 0,
          totalNumberOfReceipients: actualRecipientIds.length,
          status: 'succeeded',
          createdBy: userInfo?.id,
          broadcast_template_params: reqBody?.template_params,
          project_id: reqBody.project_id,
          total_remaining_recipients_count: actualRecipientIds.length,
        },
      })
      for (const recipient of actualRecipientIds) {
        //Initialisations
        let failCheck = false
        let createNewUser = false
        let existingUserId = null
        let chatExists = false
        let searchExistCandidate = null

        const projectInfo = await prisma.$queryRaw`select * from projects where id= ${reqBody.project_id}`

        const candidateInfo = await prisma.$queryRaw`select * from contacts as c where c.id =${recipient}`

        const chatInfo = await prisma.$queryRaw`select * from chats as c where c."candidateId" =${candidateInfo[0]?.id}`

        if (chatInfo.length > 0) {
          chatExists = true
        }

        const getValidatedPhoneNumber = await validate_phone_number(
          candidateInfo[0].phoneNumber,
          candidateInfo[0].countryCode
        )

        let whatsappAvailability = 'true'
        let headerUrl = projectInfo?.[0]?.headerUrl
        let errorLogMessage = null

        let newChat = {}

        let name = ''
        if (candidateInfo[0]?.name?.replace(/^\s+|\s+$/gm, '') === '') {
          name = getValidatedPhoneNumber.phone_number_with_country_code
        } else {
          name = candidateInfo[0]?.name
        }

        const generatedChatUid = nanoid()
        const generatedWaConversationId = nanoid()

        //Template structuring procedure starts here
        let messageStructure = ``
        let headerParameters = []
        let bodyParameters = []
        let footerParameters = []
        let components = []
        let paramArray = []
        let urlChecker = false

        //parameter structure here
        if (newBroadcast) {
          for (const param of newBroadcast.broadcast_template_params) {
            if (param.name === 'url') {
              urlChecker = true
              headerUrl = param.value
            }
            paramArray.push({
              name: param.name,
              value: param.value,
            })
          }

          paramArray = [...paramArray]
        }

        // if (templateInfo[0]?.header) {
        //   let headerText = ''
        //   if (templateInfo[0]?.header.format === 'TEXT') {
        //     headerText = templateInfo[0]?.header.text

        //     // eslint-disable-next-line no-useless-escape
        //     const pattern = /[^{\{]+(?=}\})/g
        //     let extractedHeaderParams = headerText.match(pattern)

        //     if (extractedHeaderParams) {
        //       for (const templateParam of extractedHeaderParams) {
        //         for (const candidateParam of paramArray) {
        //           if (templateParam.toLowerCase() === candidateParam.name.toLowerCase()) {
        //             let headerParamObject = {
        //               type: 'text',
        //               text: candidateParam.value,
        //             }

        //             headerParameters.push(headerParamObject)
        //             headerText = headerText.replaceAll(`{{${templateParam}}}`, `${candidateParam.value}`)
        //           }
        //         }
        //       }

        //       const key = 'text'

        //       headerParameters = [...new Map(headerParameters.map((item) => [item[key], item])).values()]
        //       if (extractedHeaderParams.length != headerParameters.length) {
        //         failCheck = true
        //       }
        //     }
        //     messageStructure = `${headerText}`
        //   } else {
        //     let headerType = String(projectInfo?.[0]?.header?.format).toLowerCase()
        //     headerParameters.push({
        //       type: headerType,
        //       [headerType]:
        //         headerType === 'document'
        //           ? {
        //               link: headerUrl,
        //               filename: `${moment.utc().format('YYYYMMDDhhmmssSSSS')}${headerType}`,
        //             }
        //           : {
        //               link: headerUrl,
        //             },
        //     })
        //   }

        //   if (headerParameters.length > 0) {
        //     components.push({
        //       type: 'header',
        //       parameters: headerParameters,
        //     })
        //   }
        // }
        if (templateInfo[0].templateHeader) {
          let headerText = ''
          if (templateInfo[0].templateHeader.format === 'TEXT') {
            headerText = templateInfo[0].templateHeader.text

            // eslint-disable-next-line no-useless-escape
            const pattern = /[^{\{]+(?=}\})/g
            let extractedHeaderParams = headerText.match(pattern)

            if (extractedHeaderParams) {
              for (const templateParam of extractedHeaderParams) {
                for (const candidateParam of paramArray) {
                  if (templateParam === candidateParam.name) {
                    let headerParamObject = {
                      type: 'text',
                      text: candidateParam.value,
                    }

                    headerParameters.push(headerParamObject)
                    headerText = headerText.replaceAll(`{{${candidateParam.name}}}`, `${candidateParam.value}`)
                  }
                }
              }

              const key = 'text'

              headerParameters = [...new Map(headerParameters.map((item) => [item[key], item])).values()]

              if (extractedHeaderParams.length != headerParameters.length) {
                return sendResponse(res, 500, {
                  message: 'Insufficient Candidate Parameters',
                })
              }
            }
            messageStructure = `${headerText}`
          }

          if (templateInfo[0].templateHeader.format === 'IMAGE') {
            if (urlChecker === false) {
              headerUrl = templateInfo[0].templateHeader?.url
            }

            let imageObject = {
              type: 'image',
              image: {
                link: headerUrl,
              },
            }

            headerParameters.push(imageObject)
          }

          if (templateInfo[0].templateHeader?.format === 'VIDEO') {
            let videoObject = {
              type: 'video',
              video: {
                link: templateInfo[0].templateHeader?.url,
              },
            }

            headerParameters.push(videoObject)
          }
          if (templateInfo[0].templateHeader?.format === 'DOCUMENT') {
            let documentObject = {
              type: 'document',
              document: {
                link: templateInfo[0].templateHeader?.url,
              },
            }

            headerParameters.push(documentObject)
          }
          if (headerParameters.length > 0) {
            components.push({
              type: 'header',
              parameters: headerParameters,
            })
          }
        }

        if (templateInfo[0]?.templateBody) {
          let bodyText = templateInfo[0].templateBody.text

          // eslint-disable-next-line no-useless-escape
          const pattern = /[^{\{]+(?=}\})/g
          let extractedBodyParams = bodyText.match(pattern)
          if (extractedBodyParams) {
            for (const templateParam of extractedBodyParams) {
              for (const candidateParam of paramArray) {
                if (templateParam.toLowerCase() === candidateParam.name.toLowerCase()) {
                  let bodyParamObject = {
                    type: 'text',
                    text: candidateParam.value,
                  }
                  bodyParameters.push(bodyParamObject)
                  bodyText = bodyText.replaceAll(`{{${templateParam}}}`, `${candidateParam.value}`)

                  break
                }
              }
            }

            if (extractedBodyParams.length != bodyParameters.length) {
              failCheck = true
            }

            const key = 'text'

            bodyParameters = [...new Map(bodyParameters.map((item) => [item[key], item])).values()]

            components.push({
              type: 'body',
              parameters: bodyParameters,
            })
          }

          if (messageStructure != ``) {
            messageStructure = `${messageStructure}
                      
          ${bodyText}`
          } else {
            messageStructure = `${bodyText}`
          }
        }

        if (templateInfo[0]?.templateFooter) {
          let footerText = templateInfo[0]?.templateFooter.text

          // eslint-disable-next-line no-useless-escape
          const pattern = /[^{\{]+(?=}\})/g
          let extractedFooterParams = footerText.match(pattern)
          if (extractedFooterParams) {
            for (const templateParam of extractedFooterParams) {
              for (const candidateParam of paramArray) {
                if (templateParam.toLowerCase() === candidateParam.name.toLowerCase()) {
                  let footerParamObject = {
                    type: 'text',
                    text: candidateParam.value,
                  }

                  footerParameters.push(footerParamObject)
                  footerText = footerText.replaceAll(`{{${templateParam}}}`, `${candidateParam.value}`)
                }
              }
            }

            if (extractedFooterParams.length != footerParameters.length) {
              failCheck = true
              //console.log("failCheck6", failCheck);
            }

            components.push({
              type: 'footer',
              parameters: footerParameters,
            })
          }

          if (messageStructure != ``) {
            messageStructure = `${messageStructure}
                
          ${footerText}`
          } else {
            messageStructure = `${footerText}`
          }
        }

        if (templateInfo[0]?.buttonsType) {
          let templateButtons = templateInfo[0]?.buttonArray

          let buttonCounter = 1

          if (templateButtons || templateButtons?.length > 0) {
            for (const button of templateButtons) {
              if (messageStructure != `` && buttonCounter === 1) {
                messageStructure = `${messageStructure}
            
              ${buttonCounter}) ${button.text}`
              } else if (messageStructure != `` && buttonCounter != 1) {
                messageStructure = `${messageStructure}
              ${buttonCounter}) ${button.text}`
              } else {
                messageStructure = `${buttonCounter}) ${button.text}`
              }

              buttonCounter++
            }
          }
        }

        const projectDetails = await prisma.projects.findFirst({
          where: {
            id: reqBody.project_id,
          },
        })

        const vendorDetails = await prisma.api_vendor.findFirst()

        const auth_config = {
          headers: {
            Authorization: `Bearer ${projectDetails?.whatsappBusinessToken}`,
            'Content-Type': 'application/json',
          },
        }

        const baseURL = vendorDetails.vendorBaseUrl
        const phoneNumber_id = projectDetails.whatsappPhoneNumberId
        const api_version = vendorDetails.vendorApiVersion

        let message = null

        message = await axios
          .post(
            `${baseURL}${api_version}/${phoneNumber_id}/messages`,
            {
              messaging_product: 'whatsapp',
              to: `${getValidatedPhoneNumber.phone_number_with_country_code}`,
              type: 'template',
              template: {
                name: templateInfo[0].name,
                language: {
                  code: templateInfo[0].languageName,
                },
                components: components,
              },
            },
            auth_config
          )
          .then((result) => {
            return result
          })
          .catch((error) => {
            return error
          })

        if (message?.response?.status === 404) {
          sendResponse(res, 404, {
            message: message?.response?.data?.error?.error_data?.details,
          })
          throw new Error(message?.response?.data?.error?.error_data?.details)
        }

        if (message?.response?.data?.error?.code === 100) {
          errorLogMessage = message?.response?.data?.error
          whatsappAvailability = 'false'
          failCheck = true
        } else if (message?.data?.messages?.[0]?.id) {
          whatsappAvailability = 'true'
        }

        if (!chatExists) {
          newChat = await prisma.chats.create({
            data: {
              candidateId: candidateInfo[0]?.id,
              chatName: name,
              status: 'open',
              chatUid: generatedChatUid,
              initiatedBy: userInfo?.id,
              waConversationId: generatedWaConversationId,
              isCandidateReplied: false,
              latestMessage: messageStructure,
              latestMessageCreatedTime: new Date().toISOString(),
              lastMessageType: templateInfo[0]?.type,
              lastSendTemplateId: reqBody.templateId,
              whatsAppAvailability: whatsappAvailability,
              projectId: reqBody.project_id,
            },
          })

          if (message?.data?.messages?.[0]?.id) {
            await prisma.chat_template.create({
              data: {
                templateId: reqBody?.templateId,
                chatId: newChat?.id,
              },
            })
          }

          //Update Chat last send message details
          const chatDetailsUpdation = await prisma.chats.update({
            data: {
              latestMessage: messageStructure,
              latestMessageCreatedTime: new Date().toISOString(),
              receiverId: candidateInfo[0]?.id,
              lastMessageType: 'text',
            },
            where: {
              id: newChat?.id,
            },
          })

          const bulkMessageData = [
            {
              chatId: newChat.id,
              owner: true,
              messageText: messageStructure,
              isMessageRead: false,
              messageType: 'text',
              chatUid: newChat?.chatUid,
              templateMessageId: templateInfo[0]?.id,
              whatsappMessageId: message?.data?.messages?.[0]?.id,
              eventType: 'template',
              timeStamp: new Date().toISOString(),
              senderId: userInfo?.id,
              receiverId: candidateInfo[0]?.id,
              status: 'active',
            },
          ]

          let individualchat = await prisma.individual_chat_details.createMany({
            data: bulkMessageData,
          })

          await prisma.broadcast_recipients.create({
            data: {
              broadcastId: newBroadcast.id,
              userId: candidateInfo[0]?.id,
              recievedStatus: 'success',
              broadcast_message_uid: message?.data?.messages?.[0]?.id,
            },
          })
        } else {
          const updateChatStatus = await prisma.chats.update({
            data: {
              status: 'open',
              latestMessage: messageStructure,
              latestMessageCreatedTime: new Date().toISOString(),
              receiverId: candidateInfo[0]?.id,
              lastMessageType: 'text',
              lastSendTemplateId: templateInfo[0]?.id,
              whatsAppAvailability: 'true',
              isCandidateReplied: true,
            },
            where: {
              chatUid: chatInfo[0]?.chatUid,
              candidateId: candidateInfo[0]?.id,
            },
          })

          const bulkData = [
            {
              chatId: updateChatStatus?.id,
              owner: true,
              messageText: messageStructure,
              whatsappMessageId: message?.data?.messages?.[0]?.id,
              isMessageRead: false,
              messageType: 'text',
              chatUid: updateChatStatus?.chatUid,
              eventType: 'template',
              senderId: userInfo?.id,
              receiverId: candidateInfo[0].id,
              status: 'active',
              timeStamp: new Date().toISOString(),
            },
          ]

          let individualchat = await prisma.individual_chat_details.createMany({
            data: bulkData,
          })

          await prisma.broadcast_recipients.create({
            data: {
              broadcastId: newBroadcast?.id,
              userId: candidateInfo[0]?.id,
              recievedStatus: 'success',
              broadcast_message_uid: message?.data?.messages?.[0]?.id,
            },
          })
        }
      }
    })

    sendResponse(res, 200, { success: true, msg: `broadcast send successfully` })
  } catch (error) {
    console.error('Error sending broadcast:', error)
  }
}

export default handler
