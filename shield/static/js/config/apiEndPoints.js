//const baseUrl = "https://a1d2-103-153-105-16.ngrok.io";
// const baseUrl = "https://dev-shield.appblocks.com";
// const baseUrl = "https://shield.appblocks.com";
// const baseUrl = "http://localhost:8011";

const baseUrl = window.location.protocol+"//"+window.location.host;


export default {
  login: `${baseUrl}/login`,
  signup: `${baseUrl}/signup`,
  changePassword: `${baseUrl}/change-password`,
  verifyEmail: `${baseUrl}/verify-user-email`,
  passwordRecovery: `${baseUrl}/password-recovery`,
  allowAppAuthPermissions: `${baseUrl}/allow-app-permissions`,
  allowAuthPermission: `${baseUrl}/allow-permissions`,
  getAllAppPermissionsForAllow: `${baseUrl}/get-app-permissions-for-allow`,
  resendOTP: `${baseUrl}/resend-user-email-otp`,


};
