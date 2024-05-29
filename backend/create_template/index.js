import { shared } from '@appblocks/node-sdk'
import { nanoid } from 'nanoid'
import axios from 'axios'

const handler = async (event) => {
  const { req, res } = event

  const { prisma, healthCheck, getBody, sendResponse, validateBody, authenticateUser } = await shared.getShared()

  try {
    // health check
    if (healthCheck(req, res)) return
    const reqBody = await getBody(req)

    await validateBody(reqBody, 'createTemplateSchema')

    const userInfo = await authenticateUser(req)

    if (userInfo.error) {
      sendResponse(res, 400, { success: false, msg: userInfo.error });
      return;
    }

    const result = await prisma.$transaction(async (tx) => {
      let headerObject = null
      let footerObject = null
      let bodyObject = null
      let buttonArray = []
      let modifiedComponentsArray = []
      let headerUrl = null

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

      const checkExistingTemplateWithName = await tx.templates.findFirst({
        where: {
          name: reqBody.name,
        },
      })

      if (checkExistingTemplateWithName) {
        return sendResponse(res, 400, {
          message: 'A template with the same name already exists. Please try with a different name!',
        })
      }

      try {
        let newClientTemplateMessage = await tx.templates.create({
          data: {
            id: nanoid(),
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

                csvData = {
                  ...csvData,
                  [`header_${param}`]: '',
                }

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
        const businessId = projectDetails.whatsappBusinessId
        let templateCreationResponse = null

        templateCreationResponse = await axios
          .post(`${baseURL}${vendorDetails.vendorApiVersion}/${businessId}/message_templates`, reqData, auth_config)
          .then((result) => {
            return result
          })
          .catch((error) => {
            return error
          })
        let templateCreationError = false
        let errorMessage = null

        if (!templateCreationResponse?.data?.id) {
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

    if (result?.template_uid || result.error === false) {
      await prisma.templates.update({
        where: {
          id: result.id,
        },
        data: {
          templateUuid: result.template_uid,
          header_url: result.header_url,
          status: result?.template_status,
        },
      })

      return sendResponse(res, 200, {
        message: 'Template Created Successfully',
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
    console.error('Error creating template:', error)
  }
}

export default handler
