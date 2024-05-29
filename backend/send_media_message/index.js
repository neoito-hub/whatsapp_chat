import { shared } from '@appblocks/node-sdk'
import axios from 'axios'
import Busboy from 'busboy'
import path from 'path'
import { default as FormData } from 'form-data'
import fs from 'fs'
const __dirname = path.resolve(path.dirname(''))

const handler = async (event) => {
  const { req, res } = event

  const { prisma, healthCheck, getBody, sendResponse, validateBody, authenticateUser } = await shared.getShared()

  try {
    // health check
    if (healthCheck(req, res)) return

    const userInfo = await authenticateUser(req);

    if (userInfo.error) {
      sendResponse(res, 400, { success: false, msg: userInfo.error })
      return
    }

    let fileName = ''
    let caption = ''
    let contactId = ''
    let project_id = ''
    let saveTo = ''
    let fileSize = ''

    // eslint-disable-next-line no-unused-vars
    await new Promise((resolve) => {
      var busboy = Busboy({ headers: req.headers })

      busboy.on(
        'file',
        // eslint-disable-next-line no-unused-vars
        function (fieldname, file, filename, encoding, mimetype) {
          file.on('data', (data) => {
            // count the bytes sent to the chunck
            fileSize += data.length
            // ...
          })
          saveTo = path.join(__dirname, filename.filename)
          fileName = filename.filename
          file.pipe(fs.createWriteStream(saveTo))
        }
      )

      busboy.on('field', (fieldname, value) => {
        if (fieldname === 'caption') {
          caption = value
        } else if (fieldname === 'contact_id') {
          contactId = value
        } else if (fieldname === 'project_id') {
          project_id = value
        }
      })

      busboy.on('finish', function () {
        resolve()
      })

      req.pipe(busboy)
    })

    let data = null

    await new Promise((resolve) => {
      fs.readFile(saveTo, function (err, content) {
        if (!err) {
          data = content
          resolve()
        } else if (err) {
          return sendResponse(res, 500, { status: 'failed' })
        }
      })
    })

    const projectDetails = await prisma.projects.findFirst({
      where: {
        id: project_id,
      },
    })
    if (contactId == '' || project_id == '' || fileName == '') {
      sendResponse(res, 404, { success: false, msg: `Inavlid payload` })
    }

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

    const BASE_URL = vendorDetails.vendorBaseUrl
    const phoneNumber_id = projectDetails.whatsappPhoneNumberId
    const api_version = vendorDetails.vendorApiVersion

    let split = fileName.split('.')
    let extension = split[split.length - 1].toLowerCase()

    let formats = {
      jpeg: {
        contentType: `image/jpeg`,
        type: 'image',
      },
      jpg: {
        contentType: `image/jpeg`,
        type: 'image',
      },
      webp: {
        contentType: `image/webp`,
        type: 'sticker',
      },
      mp4: {
        contentType: `video/mp4`,
        type: 'video',
      },
      '3gp': {
        contentType: `video/3gpp; audio/3gpp`,
        type: 'video',
      },
      pdf: {
        contentType: 'application/pdf',
        type: 'document',
      },
      png: {
        contentType: 'image/png',
        type: 'image',
      },
      doc: {
        contentType: 'application/msword',
        type: 'document',
      },
      docx: {
        contentType: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
        type: 'document',
      },
      xls: {
        contentType: 'application/vnd.ms-excel',
        type: 'document',
      },
      xlsx: {
        contentType: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
        type: 'document',
      },
      ppt: {
        contentType: 'application/vnd.ms-powerpoint',
        type: 'document',
      },
      pptx: {
        contentType: 'application/vnd.openxmlformats-officedocument.presentationml.presentation',
        type: 'document',
      },
      txt: {
        contentType: 'text/plain',
        type: 'document',
      },
      ogg: {
        contentType: 'audio/ogg',
        type: 'audio',
      },
      aac: {
        contentType: 'audio/aac',
        type: 'audio',
      },
      mp3: {
        contentType: 'audio/mpeg',
        type: 'audio',
      },
      amr: {
        contentType: 'audio/amr',
        type: 'audio',
      },
    }

    if (!formats[`${extension}`]) {
      return sendResponse(res, 500, { message: 'Invalid Media Format' })
    }
    let url = BASE_URL
    let header = {}
    let file_data = null

    url = BASE_URL + api_version + '/' + phoneNumber_id + '/media'
    let fromData = new FormData()
    fromData.append('file', fs.createReadStream(saveTo))
    fromData.append('type', formats[`${extension}`]['contentType'])
    fromData.append('messaging_product', 'whatsapp')
    header = { ...auth_config.headers, ...fromData.getHeaders() }
    file_data = fromData

    const response = await axios.post(url, file_data, { headers: header })

    let formatBody = null
    let resp_id = null

    resp_id = response.data.id

    if (formats[`${extension}`].type === 'document') {
      formatBody = {
        id: resp_id,
        filename: fileName,
      }
    } else {
      formatBody = {
        id: resp_id,
      }
    }

    if (caption) {
      formatBody = {
        ...formatBody,
        caption: caption,
      }
    }
    const axiosBodyData = {
      messaging_product: 'whatsapp',
      recipient_type: 'individual',
      to: candidateInfo[0]?.phoneNumber,
      type: formats[`${extension}`].type,
      [formats[`${extension}`].type]: formatBody,
    }

    url = BASE_URL + api_version + '/' + phoneNumber_id + '/messages'

    let message = await axios.post(url, axiosBodyData, auth_config)

    // eslint-disable-next-line no-inner-declarations
    await new Promise((resolve, reject) => {
      fs.unlink(fileName, (err) => {
        if (err) {
          reject(err)
        } else {
          // eslint-disable-next-line no-console
          console.log('success')
          resolve()
        }
      })
    })

    await prisma.$transaction(async (prisma) => {
      await prisma.individual_chat_details.create({
        data: {
          senderId: userInfo?.id,
          receiverId: candidateInfo[0].id,
          status: 'active',
          messageText: fileName,
          owner: true,
          messageType: formats[`${extension}`].type,
          chatId: chatInfo[0].id,
          eventType: 'message',
          timeStamp: new Date().toISOString(),
          waId: candidateInfo[0].phoneNumber,
          whatsappMessageId: message?.data.messages[0]?.id,
          conversationId: chatInfo[0].waConversationId,
          senderName: userInfo?.id,
          messageStatusString: 'SENT',
          fileName: fileName,
          fileType: formats[`${extension}`].type,
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
          latestMessage: fileName,
          latestMessageCreatedTime: new Date().toISOString(),
          receiverId: candidateInfo[0].id,
          lastMessageType: formats[`${extension}`].type,
        },
      })
    })

    let responseData = {
      sender_id: userInfo?.id,
      reciever_id: candidateInfo[0].name,
      message: fileName,
    }

    sendResponse(res, 200, { success: true, msg: `message send successfully`, data: responseData })
  } catch (error) {
    console.error('Error sending message:', error)
  }
}

export default handler
