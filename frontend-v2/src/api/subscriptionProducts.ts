import { xlabSubscriptionProductsAPI } from './xlab/subscriptionProducts'

export const getActive = xlabSubscriptionProductsAPI.getActive
export const getSummary = xlabSubscriptionProductsAPI.getSummary
export const getProgress = xlabSubscriptionProductsAPI.getProgress

export const subscriptionProductsAPI = {
  getActive,
  getSummary,
  getProgress
}

export default subscriptionProductsAPI
