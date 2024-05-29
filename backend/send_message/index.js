import { shared } from '@appblocks/node-sdk'
import axios from 'axios'

const handler = async (event) => {
  const { req, res } = event

  const { prisma, healthCheck, getBody, sendResponse, validateBody, authenticateUser } = await shared.getShared()

  try {
    // health check
    if (healthCheck(req, res)) return
    const reqBody = await getBody(req)

    await validateBody(reqBody, 'sendMessageSchema')

    const userInfo = await authenticateUser(req)

    if (userInfo.error) {
      sendResponse(res, 400, { success: false, msg: userInfo.error })
      return;
    }

    let contactId = reqBody.contact_id

    const projectDetails = await prisma.projects.findFirst({
      where: {
        id: reqBody.projectId,
      },
    })

    const candidateInfo = await prisma.$queryRaw`select * from contacts as c where c."id" =${contactId} `

    if (candidateInfo.length < 0) {
      throw Error('Contact not found')
    }

    const chatInfo = await prisma.$queryRaw`select * from chats as c where c."candidateId" =${candidateInfo[0]?.id}`

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

    const msgDataCloudApi = {
      messaging_product: 'whatsapp',
      recipient_type: 'individual',
      to: candidateInfo[0]?.phoneNumber,
      type: 'text',
      text: {
        preview_url: false,
        body: reqBody.message,
      },
    }

    let message = await axios.post(`${baseURL}${api_version}/${phoneNumber_id}/messages`, msgDataCloudApi, auth_config)

    await prisma.$transaction(async (prisma) => {
      await prisma.individual_chat_details.create({
        data: {
          senderId: userInfo?.id,
          receiverId: candidateInfo[0].id,
          status: 'active',
          messageText: reqBody.message,
          owner: true,
          messageType: 'text',
          chatId: chatInfo[0].id,
          eventType: 'message',
          timeStamp: new Date().toISOString(),
          waId: candidateInfo[0].phoneNumber,
          whatsappMessageId: message?.data.messages[0]?.id,
          conversationId: chatInfo[0].waConversationId,
          senderName: userInfo?.id,
          messageStatusString: 'SENT',
          isMessageRead: false,
          chatUid: chatInfo[0].chatUid,
        },
      })

      await prisma.chats.update({
        where: {
          candidateId: candidateInfo[0].id,
        },
        data: {
          status: 'open',
          latestMessage: reqBody.message,
          latestMessageCreatedTime: new Date().toISOString(),
          receiverId: candidateInfo[0].id,
          lastMessageType: 'text',
        },
      })
    })

    let responseData = {
      sender_id: userInfo?.id,
      reciever_id: candidateInfo[0].name,
      message: reqBody?.message,
    }

    sendResponse(res, 200, { success: true, msg: `message send successfully`, data: responseData })
  } catch (error) {
    console.error('Error sending message:', error)
  }
}

export default handler
