import vine from '@vinejs/vine'

const addContactSchema = vine.object({
  name: vine.string(),
  countryCode: vine.string(),
  phoneNumber: vine.string(),
  email: vine.string(),
  address: vine.string(),
  projectId: vine.string(),
})

const paginationAndSearchSchema = vine.object({
  page: vine.number(),
  limit: vine.number(),
  search: vine.string(),
  project_id: vine.string(),
})

const editContactSchema = vine.object({
  id: vine.string(),
  name: vine.string(),
  countryCode: vine.string(),
  phoneNumber: vine.string(),
  email: vine.string(),
  address: vine.string(),
})

const createTemplateSchema = vine.object({
  name: vine.string(),
  type: vine.string(),
  categoryId: vine.string(),
  buttonType: vine.string().optional(),
  projectId: vine.string(),
  category: vine.string(),
  language: vine.string(),
  languageId: vine.string(),
  components: vine.array(vine.object({})),
})

const startNewChatSchema = vine.object({
  candidate_details: vine.object({}),
  template_message_id: vine.string(),
  project_id: vine.string(),
  template_params: vine.array(vine.object({})),
})

const deleteSchema = vine.object({
  id: vine.string(),
})

const newBroadcast = vine.object({
  name: vine.string(),
  templateId: vine.string(),
  recipients: vine.array(vine.string()),
  template_params: vine.array(vine.object({})),
  project_id : vine.string()
})

const sendMessageSchema = vine.object({
  contact_id: vine.string(),
  message: vine.string(),
  project_id : vine.string()
})

const chatlistSchema = vine.object({
  page: vine.number(),
  limit: vine.number(),
  search: vine.string(),
  project_id: vine.string(),
})

const chatHistorySchema = vine.object({
  page: vine.number(),
  limit: vine.number(),
  chat_id: vine.string(),
})

export default {
  addContactSchema,
  paginationAndSearchSchema,
  editContactSchema,
  deleteSchema,
  createTemplateSchema,
  startNewChatSchema,
  newBroadcast,
  sendMessageSchema,
  chatlistSchema,
  chatHistorySchema
}
