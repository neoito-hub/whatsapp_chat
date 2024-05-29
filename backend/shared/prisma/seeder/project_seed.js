async function addProject(prisma) {
  await prisma.projects.create({
    data: {
      id: '1',
      projectUuid: '569cf0e7-1acf-4ec1-b57a-0519e876c9ea',
      name: 'Sample project',
      webhookUrl: 'YOUR_WEBHOOK_URL',
      channelName: 'YOUR_CENTRIFUGO_CHANNEL_NAME',
      whatsappBusinessId: 'YOUR_WHATSAPP_BUSINESS_ID',
      whatsappPhoneNumberId: 'YOUR_WHATSAPP_PHONENUMBER_ID',
      whatsappBusinessToken: 'YOUR_WHATSAPP_BUSINESS_TOKEN',
      webhookVerifyToken: 'YOUR_WEHOOK_VERIFY_TOKEN',
    },
  })

  console.log('Data seeded successfully')
}

export default addProject
