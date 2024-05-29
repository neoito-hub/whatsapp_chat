import { shared } from '@appblocks/node-sdk'
import { nanoid } from 'nanoid'
import axios from 'axios'

const handler = async (event) => {
  const { req, res } = event

  const { prisma, healthCheck, getBody, sendResponse, validateBody, validate_phone_number, authenticateUser } =
    await shared.getShared()

  try {
    // health check
    if (healthCheck(req, res)) return
    const reqBody = await getBody(req)

    await validateBody(reqBody, 'startNewChatSchema')

    const userInfo = await authenticateUser(req)

    if (userInfo.error) {
      sendResponse(res, 400, { success: false, msg: userInfo.error })
      return
    }

    let chatExists = false
    let whatsappAvailability = 'true'

    const getValidatedPhoneNumber = await validate_phone_number(
      reqBody.candidate_details.phone_number,
      reqBody.candidate_details.country_code
    )

    const projectInfo = await prisma.$queryRaw`select * from projects where id= ${reqBody.project_id}`

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
    t.header_url,
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
  WHERE t.id = ${reqBody.template_message_id}

`

    const candidateInfo =
      await prisma.$queryRaw`select * from contacts as c where c."phoneNumber" =${reqBody?.candidate_details?.phone_number}`

    const chatInfo = await prisma.$queryRaw`select * from chats as c where c."candidateId" =${candidateInfo[0]?.id}`

    if (chatInfo.length > 0) {
      chatExists = true
    }

    if (!candidateInfo[0].id) {
      return sendResponse(res, 404, {
        message: 'Contact not found',
      })
    }

    let contactExist = true
    if (!contactExist) {
      // new contact and new chat
      let newUser = false
      let newChat = {}
      let newContact = null

      await prisma.$transaction(async (prisma) => {
        if (!candidateInfo.length > 0) {
          newUser = true
          newContact = await prisma.contacts.create({
            data: {
              id: nanoid(),
              status: 'active',
              countryCode: reqBody.candidate_details.country_code,
              phoneNumber: reqBody.candidate_details.phone_number,
            },
          })

          const parameters = [
            {
              contactId: newContact.id,
              name: 'name',
              value: getValidatedPhoneNumber.phone_number_with_country_code,
              status: 'active',
            },
            {
              contactId: newContact.id,
              name: 'phone',
              value: getValidatedPhoneNumber.phone_number_with_country_code,
              status: 'active',
            },
          ]

          await prisma.candidate_custom_parameters.createMany({
            data: parameters,
          })
        }
        // let today = new Date();

        const generatedChatUid = nanoid()
        const generatedWaConversationId = nanoid()

        let messageStructure = ``
        let headerParameters = []
        let bodyParameters = []
        let footerParameters = []
        let components = []
        let paramArray = []
        let headerUrl = null
        let urlChecker = false

        if (reqBody?.template_params?.length > 0) {
          for (const param of reqBody.template_params) {
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
              headerUrl = templateInfo[0].templateHeader?.headerUrl
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
                link: templateInfo[0].templateHeader?.headerUrl,
              },
            }

            headerParameters.push(videoObject)
          }
          if (templateInfo[0].templateHeader?.format === 'DOCUMENT') {
            let documentObject = {
              type: 'document',
              document: {
                link: templateInfo[0].templateHeader?.headerUrl,
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

        if (templateInfo[0].templateBody) {
          let bodyText = templateInfo[0].templateBody.text

          // eslint-disable-next-line no-useless-escape
          const pattern = /[^{\{]+(?=}\})/g
          let extractedBodyParams = bodyText.match(pattern)
          if (extractedBodyParams) {
            for (const templateParam of extractedBodyParams) {
              for (const candidateParam of paramArray) {
                if (templateParam === candidateParam.name) {
                  let bodyParamObject = {
                    type: 'text',
                    text: candidateParam.value,
                  }
                  bodyParameters.push(bodyParamObject)
                  bodyText = bodyText.replaceAll(`{{${candidateParam.name}}}`, `${candidateParam.value}`)

                  break
                }
              }
            }

            if (extractedBodyParams.length != bodyParameters.length) {
              return sendResponse(res, 500, {
                message: 'Insufficient Candidate Parameters',
              })
            }

            const key = 'text'

            bodyParameters = [...new Map(bodyParameters.map((item) => [item[key], item])).values()]

            components.push({
              type: 'body',
              parameters: bodyParameters,
            })
          }

          if (messageStructure != ``) {
            messageStructure = `${messageStructure}${bodyText}`
          } else {
            messageStructure = `${bodyText}`
          }
        }

        if (templateInfo[0].templateFooter) {
          let footerText = templateInfo[0].templateFooter.text

          // eslint-disable-next-line no-useless-escape
          const pattern = /[^{\{]+(?=}\})/g
          let extractedFooterParams = footerText.match(pattern)
          if (extractedFooterParams) {
            for (const templateParam of extractedFooterParams) {
              for (const candidateParam of paramArray) {
                if (templateParam === candidateParam.name) {
                  let footerParamObject = {
                    type: 'text',
                    text: candidateParam.value,
                  }

                  footerParameters.push(footerParamObject)
                  footerText = footerText.replaceAll(`{{${candidateParam.name}}}`, `${candidateParam.value}`)
                }
              }
            }

            if (extractedFooterParams.length != footerParameters.length) {
              return sendResponse(res, 500, {
                message: 'Insufficient Candidate Parameters',
              })
            }

            components.push({
              type: 'footer',
              parameters: footerParameters,
            })
          }

          if (messageStructure != ``) {
            messageStructure = `${messageStructure}${footerText}`
          } else {
            messageStructure = `${footerText}`
          }
        }

        if (templateInfo[0]?.buttonsType) {
          let templateButtons = templateInfo[0]?.buttonArray

          let buttonCounter = 1
          if (templateButtons?.length > 0) {
            for (const button of templateButtons) {
              if (messageStructure != `` && buttonCounter === 1) {
                messageStructure = `${messageStructure}${buttonCounter}) ${button.text}`
              } else if (messageStructure != `` && buttonCounter != 1) {
                messageStructure = `${messageStructure}${buttonCounter}) ${button.text}`
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

        let message = await axios
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
          whatsappAvailability = 'false'
          sendResponse(res, 400, {
            message: message?.response?.data?.error?.error_data?.details,
          })
        } else if (message?.data?.messages?.[0]?.id) {
          whatsappAvailability = 'true'
        }

        let lastMessageType = 'text'
        let latestMessageCreatedTime = new Date().toISOString()

        if (whatsappAvailability === 'false') {
          messageStructure = ''
          lastMessageType = null
          latestMessageCreatedTime = null
        }

        newChat = await prisma.chats.create({
          data: {
            candidateId: newUser ? newContact.id : candidateInfo[0].id,
            chatName: newContact?.name,
            status: 'open',
            chatUid: generatedChatUid,
            initiatedBy: userInfo?.id,
            waConversationId: generatedWaConversationId,
            isCandidateReplied: false,
            latestMessage: messageStructure,
            latestMessageCreatedTime: latestMessageCreatedTime,
            receiverId: newUser ? newContact.id : candidateInfo[0].id,
            lastMessageType: lastMessageType,
            lastSendTemplateId: reqBody.template_message_id,
            whatsAppAvailability: whatsappAvailability,
            projectId: reqBody.project_id,
          },
        })

        await prisma.chat_template.create({
          data: {
            templateId: reqBody.template_message_id,
            chatId: newChat.id,
          },
        })

        const bulkMessageData = [
          {
            chatId: newChat.id,
            owner: true,
            messageText: messageStructure,
            isMessageRead: false,
            messageType: 'text',
            chatUid: newChat.chatUid,
            templateMessageId: templateInfo[0]?.id,
            eventType: 'template',
            senderId: userInfo?.id,
            receiverId: newUser ? newContact.id : candidateInfo[0].id,
            status: 'active',
          },
        ]

        await prisma.individual_chat_details.createMany({
          data: bulkMessageData,
        })
        return sendResponse(res, 200, {
          message: 'Successfully initiated chat',
        })
      })
    } else {
      //exsisting user chat handling
      let messageStructure = ``
      let headerParameters = []
      let bodyParameters = []
      let footerParameters = []
      let components = []

      let paramArray = []
      let headerUrl = null
      let nameObject = null
      let phoneObject = null
      let urlChecker = false
      let newChat = {}
      let latestMessageCreatedTime = new Date().toISOString()
      let lastMessageType = 'text'

      const generatedChatUid = nanoid()
      const generatedWaConversationId = nanoid()

      let candidateDetails = await prisma.$queryRaw`
      select ccp.id,ccp."contactId",ccp.name as param_name,ccp.value as param_value,c.status,c.id,
      c.name as contact_name,c."countryCode",c."phoneNumber",c.email
      from contacts as c
      inner join candidate_custom_parameters as ccp on  ccp."contactId"=c.id
      where c.id =${candidateInfo[0]?.id}`

      let candidateChatInfo = await prisma.$queryRaw`
       select * from chats as c
       where c."candidateId" =${candidateInfo[0]?.id}`

      if (reqBody?.template_params?.length > 0) {
        for (const param of reqBody.template_params) {
          if (param.name === 'url') {
            urlChecker = true
            headerUrl = param.value
          }

          paramArray.push({
            name: param.name,
            value: param.value,
          })
        }

        for (const candidate of candidateDetails) {
          if (candidate.param_name === 'name') {
            nameObject = {
              name: candidate.param_name,
              value: candidate.param_value,
            }
          }

          if (candidate.param_name === 'phone') {
            phoneObject = {
              name: candidate.param_name,
              value: candidate.param_value,
            }
          }

          if (nameObject !== null && phoneObject !== null) {
            break
          }
        }

        paramArray = [...paramArray, nameObject, phoneObject]
      }

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

      if (templateInfo[0].templateBody) {
        let bodyText = templateInfo[0].templateBody.text

        // eslint-disable-next-line no-useless-escape
        const pattern = /[^{\{]+(?=}\})/g
        let extractedBodyParams = bodyText.match(pattern)
        if (extractedBodyParams) {
          for (const templateParam of extractedBodyParams) {
            for (const candidateParam of paramArray) {
              if (templateParam === candidateParam.name) {
                let bodyParamObject = {
                  type: 'text',
                  text: candidateParam.value,
                }
                bodyParameters.push(bodyParamObject)
                bodyText = bodyText.replaceAll(`{{${candidateParam.name}}}`, `${candidateParam.value}`)

                break
              }
            }
          }

          if (extractedBodyParams.length != bodyParameters.length) {
            return sendResponse(res, 500, {
              message: 'Insufficient Candidate Parameters',
            })
          }

          const key = 'text'

          bodyParameters = [...new Map(bodyParameters.map((item) => [item[key], item])).values()]

          components.push({
            type: 'body',
            parameters: bodyParameters,
          })
        }

        if (messageStructure != ``) {
          messageStructure = `${messageStructure}${bodyText}`
        } else {
          messageStructure = `${bodyText}`
        }
      }

      if (templateInfo[0].templateFooter) {
        let footerText = templateInfo[0].templateFooter.text

        // eslint-disable-next-line no-useless-escape
        const pattern = /[^{\{]+(?=}\})/g
        let extractedFooterParams = footerText.match(pattern)
        if (extractedFooterParams) {
          for (const templateParam of extractedFooterParams) {
            for (const candidateParam of paramArray) {
              if (templateParam === candidateParam.name) {
                let footerParamObject = {
                  type: 'text',
                  text: candidateParam.value,
                }

                footerParameters.push(footerParamObject)
                footerText = footerText.replaceAll(`{{${candidateParam.name}}}`, `${candidateParam.value}`)
              }
            }
          }

          if (extractedFooterParams.length != footerParameters.length) {
            return sendResponse(res, 500, {
              message: 'Insufficient Candidate Parameters',
            })
          }

          components.push({
            type: 'footer',
            parameters: footerParameters,
          })
        }

        if (messageStructure != ``) {
          messageStructure = `${messageStructure}${footerText}`
        } else {
          messageStructure = `${footerText}`
        }
      }

      if (templateInfo[0].buttonsType) {
        let templateButtons = templateInfo[0]?.buttonArray

        let buttonCounter = 1
        if (templateButtons?.length > 0) {
          for (const button of templateButtons) {
            if (messageStructure != `` && buttonCounter === 1) {
              messageStructure = `${messageStructure}${buttonCounter}) ${button.text}`
            } else if (messageStructure != `` && buttonCounter != 1) {
              messageStructure = `${messageStructure}${buttonCounter}) ${button.text}`
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

      let message = await axios
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
        latestMessageCreatedTime = null
        lastMessageType = null
        await prisma.chats.update({
          data: {
            whatsAppAvailability: 'false',
          },

          where: {
            candidateId: candidateInfo[0]?.id,
            id: candidateChatInfo[0]?.id,
          },
        })

        //https://developers.facebook.com/docs/whatsapp/cloud-api/support/error-codes
        sendResponse(res, 400, {
          message: message?.response?.data?.error?.error_data?.details ?? 'Recipient is not a valid WhatsApp user',
        })
        throw new Error(message?.response?.data?.error?.error_data?.details ?? 'Recipient is not a valid WhatsApp user')
      }

      if (message?.data?.messages?.[0]?.id) {
        prisma.$transaction(async (prisma) => {
          if (!chatExists) {
            newChat = await prisma.chats.create({
              data: {
                candidateId: candidateDetails[0].id,
                chatName: candidateDetails[0]?.contact_name,
                status: 'open',
                chatUid: generatedChatUid,
                initiatedBy: userInfo?.id,
                waConversationId: generatedWaConversationId,
                isCandidateReplied: false,
                latestMessage: messageStructure,
                latestMessageCreatedTime: latestMessageCreatedTime,
                receiverId: candidateDetails[0].id,
                lastMessageType: lastMessageType,
                lastSendTemplateId: reqBody.template_message_id,
                whatsAppAvailability: whatsappAvailability,
                projectId: reqBody.project_id,
              },
            })

            await prisma.chat_template.create({
              data: {
                templateId: reqBody.template_message_id,
                chatId: newChat.id,
              },
            })

            const bulkMessageData = [
              {
                chatId: newChat.id,
                owner: true,
                messageText: messageStructure,
                isMessageRead: false,
                messageType: 'text',
                chatUid: newChat.chatUid,
                templateMessageId: templateInfo[0]?.id,
                eventType: 'template',
                senderId: userInfo?.id,
                receiverId: candidateDetails[0].id,
                status: 'active',
              },
            ]

            await prisma.individual_chat_details.createMany({
              data: bulkMessageData,
            })
            return sendResponse(res, 200, {
              message: 'Successfully initiated chat',
            })
          } else {
            if (candidateChatInfo?.latest_message === '') {
              await prisma.chat_template.create({
                data: {
                  templateId: reqBody.template_message_id,
                  chatId: candidateChatInfo[0]?.id,
                },
              })
            } else {
              let userTemplateSentCheckData = await prisma.$queryRaw`
           select id from chat_template where "templateId"= ${templateInfo[0]?.id} and "chatId" =${candidateChatInfo[0].id}`

              if (!userTemplateSentCheckData?.[0]?.id) {
                //If they have not received this before store the template sending history and update last send template history
                await prisma.chat_template.create({
                  data: {
                    templateId: reqBody.template_message_id,
                    chatId: candidateChatInfo[0]?.id,
                  },
                })
              }
            }

            const updateChatStatus = await prisma.chats.update({
              data: {
                status: 'open',
                latestMessage: messageStructure,
                latestMessageCreatedTime: new Date().toISOString(),
                receiverId: candidateDetails[0].id,
                lastMessageType: 'text',
                projectId: reqBody?.project_id,
                lastSendTemplateId: reqBody.template_message_id,
                whatsAppAvailability: 'true',
                isCandidateReplied: candidateChatInfo[0].isCandidateReplied,
              },
              where: {
                candidateId: candidateDetails[0].id,
              },
            })

            const bulkData = [
              {
                chatId: candidateChatInfo[0].id,
                owner: true,
                messageText: messageStructure,
                isMessageRead: false,
                messageType: 'text',
                chatUid: updateChatStatus?.chatUid,
                templateMessageId: templateInfo[0]?.id,
                eventType: 'template',
                senderId: userInfo?.id,
                receiverId: candidateDetails[0].id,
                status: 'active',
              },
            ]

            await prisma.individual_chat_details.createMany({
              data: bulkData,
            })

            return sendResponse(res, 200, {
              message: 'message sent, chat initiated successfully',
            })
          }
        })
      }
    }
  } catch (error) {
    console.error('Error initiating chat:', error)
  }
}

export default handler
