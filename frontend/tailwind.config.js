const config = {
  content: ['./src/**/*.{js,jsx,ts,tsx}'],
  theme: {
    extend: {
      animation: {
        bounce200: 'bounce 1s infinite 400ms',
        bounce400: 'bounce 1s infinite 800ms',
      },
      screens: {
        md: '769px',
        'md-lt': {
          max: '768px',
        },
        'lg-lt': {
          max: '1023px',
        },
        xxs: {
          max: '420px',
        },
        '3xl': {
          min: '1820px',
        },
      },
      fontSize: {
        'ab-sm': ['13px', 'normal'],
        'ab-base': ['15px', 'normal'],
        'ab-3xl': ['28px', 'normal'],
      },
      colors: {
        primary: '#5E5EDD',
        secondary: '#D453B6',
        warning: '#FFF0F0',
        'ab-disabled': '#C2C2C2',
        'primary-light': '#F2EBFF',
        ab: {
          red: '#EB0000',
          green: '#01944C',
          black: '#484848',
          yellow: '#FFE01B',
          orange: '#FFA500',
          'disabled-yellow': '#FCF5CD',
          gray: {
            light: '#F8F8F8',
            dark: '#DDDDDD',
            medium: '#E0E0E0',
            bold: '#6A737D',
          },
        },
      },
      boxShadow: {
        box: '0px 8px 8px rgba(0, 0, 0, 0.024)',
        'box-md': '0px 0px 8px rgba(0, 0, 0, 0.12)',
      },
      // Add your custom styles
      components: {
        '.btn-secondary': {
          backgroundColor: '#D453B6',
          color: 'white',
          padding: '0.375rem 1.5rem',
          borderRadius: '0.375rem',
          fontWeight: 'bold',
          textAlign: 'center',
          cursor: 'pointer',
          transition: 'all 0.3s',
          '&:hover': {
            backgroundColor: 'rgba(0, 0, 0, 0.8)',
          },
        },
        '.btn-primary': {
          backgroundColor: '#5E5EDD',
          color: 'white',
          padding: '0.375rem 1.5rem',
          borderRadius: '0.375rem',
          fontWeight: 'bold',
          textAlign: 'center',
          cursor: 'pointer',
          transition: 'all 0.3s',
          '&:hover': {
            backgroundColor: 'rgba(0, 0, 0, 0.8)',
          },
        },
        '.btn-default': {
          color: 'white',
          padding: '0.375rem 1.5rem',
          borderRadius: '0.375rem',
          fontWeight: 'bold',
          textAlign: 'center',
          cursor: 'pointer',
          transition: 'opacity 0.3s',
          '&:hover': {
            backgroundColor: 'rgba(0, 0, 0, 0.8)',
          },
        },
        '.nav-menu-wrapper': {
          width: '16rem',
          position: 'fixed',
          height: 'calc(100vh - 4rem)',
          overflowY: 'auto',
          overflowX: 'hidden',
          paddingLeft: '0.75rem',
          borderLeft: '1px solid #yourBorderColor',
          backgroundColor: '#110D17',
          transition: 'all 0.5s linear',
        },
        '.info-box': {
          display: 'flex',
          alignItems: 'center',
          backgroundColor: '#yourInfoBoxColor',
          flexShrink: 0,
          border: '1px solid #yourBorderColor',
          borderRadius: '9999px',
          lineHeight: '1',
        },
        '.ab-tooltip': {
          fontWeight: 'bold',
          fontSize: '0.75rem',
          borderRadius: '0.375rem',
          backgroundColor: '#yourTooltipBgColor',
        },
        '.radio-icon': {
          border: '1px solid #yourBorderColor',
          float: 'left',
          height: '1.25rem',
          width: '1.25rem',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          flexShrink: 0,
          borderRadius: '9999px',
          backgroundColor: 'white',
          '&:after': {
            content: '""',
            borderRadius: '50%',
            height: '0.75rem',
            width: '0.75rem',
            backgroundColor: 'transparent',
          },
          '&.peer-checked:after': {
            backgroundColor: '#yourCheckedColor',
          },
        },
      },
    },
  },
  plugins: [],
}
export default config
