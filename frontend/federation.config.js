const getFederationConfig = () => ({
  name: 'frontend',
  filename: 'remoteEntry.js',
  exposes: {
    './frontend': './src/App.js',
  },
  shared: {
    react: {
      import: 'react', // the "react" package will be used a provided and fallback module
      shareKey: 'react', // under this name the shared module will be placed in the share scope
      shareScope: 'default', // share scope with this name will be used
      singleton: true, // only a single version of the shared module is allowed
      version: '^17.0.2',
    },
    'react-dom': {
      import: 'react-dom', // the "react" package will be used a provided and fallback module
      shareKey: 'react-dom', // under this name the shared module will be placed in the share scope
      shareScope: 'default', // share scope with this name will be used
      version: '^17.0.2',
      singleton: true,
    },
    'react-redux': {
      import: 'react-redux', // the "react" package will be used a provided and fallback module
      shareKey: 'react-redux', // under this name the shared module will be placed in the share scope
      shareScope: 'default', // share scope with this name will be used
      version: '^7.2.5',
      singleton: true,
    },
    'react-router-dom': {
      import: 'react-router-dom', // the "react" package will be used a provided and fallback module
      shareKey: 'react-router-dom', // under this name the shared module will be placed in the share scope
      shareScope: 'default', // share scope with this name will be used
      singleton: true, // only a single version of the shared module is allowed
      version: '^5.2.0',
    },
    '@appblocks/js-sdk': {
      import: '@appblocks/js-sdk',
      shareKey: '@appblocks/js-sdk',
      shareScope: 'default',
      singleton: true,
      version: '^0.0.9',
    },
  },
})

export default getFederationConfig
