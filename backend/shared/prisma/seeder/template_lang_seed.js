async function addTemplateLang(prisma) {
  await prisma.template_languages.create({
    data: {
      id: '1',
      name: 'English(UK)',
      code: 'en_GB',
    },
  })

  console.log('Data seeded successfully')
}

export default addTemplateLang