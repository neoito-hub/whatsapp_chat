import { shared } from '@appblocks/node-sdk'
import axios from 'axios'
import moment from 'moment'

const handler = async (event) => {
  const { req, res } = event

  const { prisma, healthCheck, getBody, sendResponse, validateBody, authenticateUser } = await shared.getShared()

  try {
    // health check
    if (healthCheck(req, res)) return
    const reqBody = await getBody(req)

    await validateBody(reqBody, 'editTemplateSchema')

    const userInfo = await authenticateUser(req)

    if (userInfo.error) {
      sendResponse(res, 400, { success: false, msg: userInfo.error })
      return
    }
    const templateDetails = await prisma.templates.findFirst({
      where: {
        id: reqBody.template_id,
      },
    })
    if (templateDetails == null) {
      return sendResponse(res, 400, {
        message: 'Failed! template not found',
      })
    }

    const result = await prisma.$transaction(async (tx) => {
      let headerObject = null
      let footerObject = null
      let bodyObject = null
      let buttonArray = []
      let modifiedComponentsArray = []
      let headerUrl = null

      if (
        templateDetails.name !== reqBody.name ||
        templateDetails.languageName !== reqBody.language ||
        templateDetails.templateType !== reqBody.type ||
        templateDetails.languageId !== reqBody.languageId ||
        templateDetails.categoryId !== reqBody.categoryId
      ) {
        return sendResponse(res, 400, {
          message: 'Failed! only category or components can be edited',
        })
      }

      if (templateDetails.updatedAt && templateDetails.status === 'APPROVED') {
        const lastEditedTime = moment(templateDetails.updatedAt)
        const currentTime = moment(new Date())
        const diffInTime = moment.duration(currentTime.diff(lastEditedTime))
        if (diffInTime.asHours() < 24) {
          return sendResponse(res, 400, {
            message: 'Failed! Approved template can be edited only one time in a 24 hour window',
          })
        } else {
          templateDetails.updatedAt = new Date()
          await templateDetails.save()
        }
      }

      // eslint-disable-next-line no-useless-escape
      const pattern = /[^{\{]+(?=}\})/g

      for (const component of reqBody.components) {
        switch (component.type) {
          case 'BODY': {
            bodyObject = component
            break
          }
          case 'HEADER': {
            headerObject = component
            break
          }
          case 'FOOTER': {
            footerObject = component
            break
          }
          case 'BUTTONS': {
            buttonArray = component.buttons
            break
          }
        }
      }

      try {
        let newClientTemplateMessage = await tx.templates.update({
          where: { id: reqBody.template_id },
          data: {
            status: 'pending',
            name: reqBody?.name.toLowerCase(),
            categoryName: reqBody?.category,
            languageName: reqBody?.language,
            templateType: reqBody?.type,
            templateHeader: headerObject,
            templateBody: bodyObject,
            templateFooter: footerObject,
            type: 'template',
            buttonsType: reqBody.buttonType ? reqBody.buttonType : '',
            category: { connect: { id: reqBody?.categoryId } },
            language: { connect: { id: reqBody?.languageId } },
            project: { connect: { id: reqBody?.projectId } },
          },
        })

        if (buttonArray.length > 0) {
          await tx.template_message_buttons.delete({
            where: { templateMessageId: reqBody.template_id }, // Assuming `reqBody.id` contains the ID of the template to delete
          })

          await tx.template_message_buttons.create({
            data: {
              templateMessageId: newClientTemplateMessage.id,
              buttonArray: buttonArray,
            },
          })
        }

        if (headerObject != null) {
          if (
            headerObject.format === 'IMAGE' ||
            headerObject.format === 'DOCUMENT' ||
            headerObject.format === 'VIDEO'
          ) {
            headerUrl = headerObject?.url
            delete headerObject.url
            delete headerObject.text

            headerObject = {
              ...headerObject,
              text: null,
              buttons: null,
            }

            modifiedComponentsArray.push(headerObject)
          } else {
            let stringifiedHeaderObject = JSON.stringify(headerObject)
            let extractedHeaderParams = stringifiedHeaderObject.match(pattern)
            extractedHeaderParams = [...new Set(extractedHeaderParams)]
            let paramIncrementor = 1
            let headerParams = []
            let exampleArray = []

            if (extractedHeaderParams) {
              for (const param of extractedHeaderParams) {
                exampleArray.push('header')
                stringifiedHeaderObject = stringifiedHeaderObject.replaceAll(`{{${param}}}`, `{{${paramIncrementor}}}`)

                headerParams.push({
                  templateMessageId: newClientTemplateMessage.id,
                  paramName: param,
                  paramValue: null,
                  paramOrder: paramIncrementor,
                  paramType: 'header',
                })

                paramIncrementor++
              }
            }

            if (headerParams.length != 0) {
              await tx.template_message_custom_parameters.deleteMany({
                where: {
                  templateMessageId: reqBody.template_id,
                },
              })
              await tx.template_message_custom_parameters.createMany({
                data: headerParams,
              })
            }

            let parsedHeaderObject = JSON.parse(stringifiedHeaderObject)

            if (extractedHeaderParams?.length > 0) {
              parsedHeaderObject = {
                ...parsedHeaderObject,
                example: {
                  header_text: exampleArray,
                },
              }
            }

            parsedHeaderObject = {
              ...parsedHeaderObject,
              buttons: null,
            }

            modifiedComponentsArray.push(parsedHeaderObject)
          }
        }

        if (bodyObject != null) {
          let stringifiedBodyObject = JSON.stringify(bodyObject)
          let extractedBodyParams = stringifiedBodyObject.match(pattern)
          extractedBodyParams = [...new Set(extractedBodyParams)]
          let paramIncrementor = 1
          let bodyParams = []
          let exampleArray = []

          if (extractedBodyParams) {
            for (const param of extractedBodyParams) {
              exampleArray.push('body')

              stringifiedBodyObject = stringifiedBodyObject.replaceAll(`{{${param}}}`, `{{${paramIncrementor}}}`)

              bodyParams.push({
                templateMessageId: newClientTemplateMessage.id,
                paramName: param,
                paramValue: null,
                paramOrder: paramIncrementor,
                paramType: 'body',
              })

              paramIncrementor++
            }
          }

          if (bodyParams.length != 0) {
            await tx.template_message_custom_parameters.deleteMany({
              where: {
                templateMessageId: reqBody.template_id,
              },
            })
            await tx.template_message_custom_parameters.createMany({
              data: bodyParams,
            })
          }

          let parsedBodyObject = JSON.parse(stringifiedBodyObject)

          if (extractedBodyParams?.length > 0) {
            parsedBodyObject = {
              ...parsedBodyObject,
              example: {
                body_text: [exampleArray],
              },
            }
          }

          parsedBodyObject = {
            ...parsedBodyObject,
            format: null,
            buttons: null,
          }

          modifiedComponentsArray.push(parsedBodyObject)
        }

        if (footerObject != null) {
          footerObject.text = footerObject.text.replace(/(\r\n|\n|\r)/gm, '')

          let stringifiedFooterObject = JSON.stringify(footerObject)
          let extractedFooterParams = stringifiedFooterObject.match(pattern)
          let paramIncrementor = 1
          let footerParams = []

          if (extractedFooterParams) {
            for (const param of extractedFooterParams) {
              stringifiedFooterObject = stringifiedFooterObject.replaceAll(`{{${param}}}`, `{{${paramIncrementor}}}`)

              footerParams.push({
                templateMessageId: newClientTemplateMessage.id,
                paramName: param,
                paramValue: null,
                paramOrder: paramIncrementor,
                paramType: 'footer',
              })

              paramIncrementor++
            }
          }

          if (footerParams.length != 0) {
            await tx.template_message_custom_parameters.deleteMany({
              where: {
                templateMessageId: reqBody.template_id,
              },
            })
            await tx.template_message_custom_parameters.createMany({
              data: footerParams,
            })
          }

          modifiedComponentsArray.push(JSON.parse(stringifiedFooterObject))
        }

        if (buttonArray.length > 0) {
          let paramIncrementor = 1
          let buttonArrayIndex = 0
          let buttonParams = []

          for (const button of buttonArray) {
            if (button.type === 'PHONE_NUMBER') {
              button.phone_number = `${button.country_code}${button.phone_number}`
              delete button.country_code
            }
            let stringifiedButtonObject = JSON.stringify(button)
            let extractedButtonParams = stringifiedButtonObject.match(pattern)
            if (extractedButtonParams?.length > 0) {
              for (const param of extractedButtonParams) {
                stringifiedButtonObject = stringifiedButtonObject.replaceAll(`{{${param}}}`, `{{${paramIncrementor}}}`)

                buttonParams.push({
                  templateMessageId: newClientTemplateMessage.id,
                  paramName: param,
                  paramValue: null,
                  paramOrder: paramIncrementor,
                  paramType: 'button',
                })

                paramIncrementor++
              }
            }

            buttonArray[buttonArrayIndex] = JSON.parse(stringifiedButtonObject)
            buttonArrayIndex++
          }

          if (buttonParams.length != 0) {
            await tx.template_message_custom_parameters.deleteMany({
              where: {
                templateMessageId: reqBody.template_id,
              },
            })
            await tx.template_message_custom_parameters.createMany({
              data: buttonParams,
            })
          }

          let buttonBody = {
            type: 'BUTTONS',
            buttons: buttonArray,
          }

          buttonBody = {
            ...buttonBody,
            format: null,
            text: null,
            example: null,
          }

          modifiedComponentsArray.push(buttonBody)
        }

        if (modifiedComponentsArray.length != 0) {
          reqBody.components = modifiedComponentsArray
        }

        delete reqBody.languageId
        delete reqBody.categoryId
        delete reqBody.type
        delete reqBody.buttonType
        delete reqBody.projectId
        reqBody.allow_category_change = true

        let reqData = { ...reqBody, name: reqBody.name.toLowerCase() }

        const projectDetails = await tx.projects.findFirst({
          where: {
            id: reqBody?.projectId,
          },
        })

        const vendorDetails = await tx.api_vendor.findFirst()

        const auth_config = {
          headers: {
            Authorization: `Bearer ${projectDetails?.whatsappBusinessToken}`,
            'Content-Type': 'application/json',
          },
        }

        const baseURL = vendorDetails.vendorBaseUrl
        let templateCreationResponse = null

        templateCreationResponse = await axios
          .post(`${baseURL}v17.0/${templateDetails.templateUuid}`, reqData, auth_config)
          .then((result) => {
            return result
          })
          .catch((error) => {
            return error
          })
        let templateCreationError = false
        let errorMessage = null

        if (!templateCreationResponse?.data?.success == true) {
          templateCreationError = true
          errorMessage = templateCreationResponse.response.data?.error?.message
        }
        return {
          id: newClientTemplateMessage?.id ?? null,
          template_uid: templateCreationResponse?.data?.id ?? null,
          template_status: templateCreationResponse?.data?.status ?? null,
          header_url: headerUrl,
          error: templateCreationError,
          error_message: errorMessage,
        }
      } catch (err) {
        console.log('----err---', err)
      }
    })

    if (result.error === false) {
      await prisma.templates.update({
        where: {
          id: reqBody.template_id,
        },
        data: {
          templateUuid: result.template_uid,
          header_url: result.header_url,
          status: 'APPROVED',
        },
      })

      return sendResponse(res, 200, {
        message: 'Template Updated Successfully',
      })
    } else {
      await prisma.template_message_custom_parameters.deleteMany({
        where: {
          templateMessageId: result.id,
        },
      })

      await prisma.template_message_buttons.deleteMany({
        where: {
          templateMessageId: result.id,
        },
      })

      await prisma.templates.deleteMany({
        where: {
          id: result.id,
        },
      })

      return sendResponse(res, 400, { message: result.error_message })
    }
  } catch (error) {
    console.error('Error updating template:', error)
  }
}

export default handler
