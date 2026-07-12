export const LOGIN_REQUIRED_EVENT='tide:login-required'
export function requestLogin(){window.dispatchEvent(new CustomEvent(LOGIN_REQUIRED_EVENT))}
