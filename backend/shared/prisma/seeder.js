import { PrismaClient } from '@prisma/client'
// import addProject from './seeder/project_seed.js'
// import addVendor from './seeder/vendor_seed.js'
// import addTemplateCategory from './seeder/template_category_seed.js'
// import addTemplateLang from './seeder/template_lang_seed.js'
const prisma = new PrismaClient()

async function main() {
  // await  addProject(prisma)
  // await  addVendor(prisma)
  // await  addTemplateCategory(prisma)
  // await  addTemplateLang(prisma)
}

main()
  .catch((e) => {
    console.log('e', e)
    process.exit(1)
  })
  .finally(async () => {
    await prisma.$disconnect()
  })
