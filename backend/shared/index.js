import prisma from './prisma/index.js'
import utils from './utils/index.js'
import 'dotenv/config'; 
import validateBody from './validations/index.js'

export default {
  ...utils,
  prisma,
  validateBody,
}
