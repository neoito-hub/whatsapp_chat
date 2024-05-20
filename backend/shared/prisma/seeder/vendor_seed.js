async function addVendor(prisma) {
  await prisma.api_vendor.create({
    data: {
      id: 1,
      vendorType: 'Cloud API',
      vendorName: 'WhatsApp Cloud API',
      vendorBaseUrl: 'https://graph.facebook.com/',
      vendorApiVersion: 'v15.0',
      status: 'active',
    },
  })

  console.log('Data seeded successfully')
}

export default addVendor
