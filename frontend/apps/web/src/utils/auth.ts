export const LOGIN_REQUIRED_EVENT='tide:login-required'
export const CREDIT_PACKAGES_EVENT='tide:credit-packages-required'
export function requestLogin(){window.dispatchEvent(new CustomEvent(LOGIN_REQUIRED_EVENT))}
