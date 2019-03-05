import AutoRefreshPrpcClient from './prpc.js';

export const prpcClient = new AutoRefreshPrpcClient(
  window.CS_env.token, window.CS_env.tokenExpiresSec);
