/* eslint-disable import/no-extraneous-dependencies */
import * as Yup from 'yup'

const CTAButtonSchema = Yup.array().of(
  Yup.object().shape({
    text: Yup.string().required('Button label is required'),
    type: Yup.string().required(),
    url: Yup.string().when('type', (type, schema) =>
      type && type[0] === 'URL'
        ? schema.required('URL is required').url('Invalid URL format')
        : schema
    ),
    country_code: Yup.string().when('type', (type, schema) =>
      type && type[0] === 'PHONE_NUMBER'
        ? schema.required('Country code is required')
        : schema
    ),
    // phone_number: Yup.string().when(
    //   ['type', 'country_code'],
    //   (values, schema) => values && values[0] === 'PHONE_NUMBER'
    //       ? schema
    //           .required('Phone number is required')
    //           .test('is-valid-phone-number', 'Invalid phone number', (num) =>
    //             validateNumber(`+${values[1]}${num}`)
    //           )
    //           .min(7, 'Phone number must be a valid number')
    //       : schema
    // ),
  })
)

const QRButtonSchema = Yup.array().of(
  Yup.object().shape({
    text: Yup.string().required('Button text is required'),
    type: Yup.string().required('Type is required'),
  })
)
const templateSchema = Yup.object().shape({
  category: Yup.string().required('Select category'),
  language: Yup.string().required('select a language'),
  language_id: Yup.string().required('select a language'),
  category_id: Yup.string().required('Select category'),
  type: Yup.string().required(),
  name: Yup.string().required('Template name is required'),
  components: Yup.array().of(
    // @ts-ignore
    // eslint-disable-next-line consistent-return
    Yup.lazy((value) => {
      if (value && value.type === 'BODY') {
        return Yup.object().shape({
          type: Yup.string().required(),
          text: Yup.string()
            .test('not-newline', 'Body text is required', (val) => val !== '\n')
            .required(),
        })
      }
      if (value && value.type === 'HEADER') {
        return Yup.object().shape({
          type: Yup.string().required(),
          format: Yup.string(),
          text: Yup.string().when(['format'], (format, schema) =>
            format && format[0] === 'TEXT'
              ? schema.required('Header text is required')
              : schema
          ),
          url: Yup.string().when('format', (format, schema) =>
            format && format[0] !== 'TEXT'
              ? schema.required('Media is required')
              : schema
          ),
          example: Yup.object().when('format', (values, schema) => {
            if (values[0] !== 'TEXT') {
              return schema.required('Example is required')
            }
            return schema
          }),
        })
      }
      if (value && value.type === 'FOOTER') {
        return Yup.object().shape({
          type: Yup.string().required(),
          text: Yup.string()
            .test(
              'not-newline',
              'Footer text is required',
              (val) => val !== '\n'
            )
            .required(),
        })
      }
      if (value && value.type === 'BUTTONS') {
        return Yup.object().shape({
          button_type: Yup.string().required(),
          type: Yup.string().required(),
          buttons:
            value.button_type === 'quick_reply'
              ? QRButtonSchema.required()
              : CTAButtonSchema.required(),
        })
      }
    })
  ),
  button_type: Yup.string().when('components', (components, schema) => {
    if (components && components.some((c) => c.type === 'BUTTONS')) {
      return schema.required(
        'Button type is required when components have buttons'
      )
    }
    return schema
  }),
})

const CsvBrdPayloadSchema = Yup.object().shape({
  name: Yup.string().required('Broadcast name is required'),
  templateId: Yup.string().required('Choose a template'),
  is_scheduled_for_later: Yup.boolean().required(),
  dynamic_data_binding: Yup.boolean().required(),
  csvData: Yup.mixed().required(),
})

export { CsvBrdPayloadSchema, templateSchema }
