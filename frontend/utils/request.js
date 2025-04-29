/**
 * 请求工具类
 */

// API基础URL
const BASE_URL = 'http://localhost:8080';

// 请求拦截器
const requestInterceptor = (config) => {
  // 获取token
  const token = uni.getStorageSync('token');
  if (token) {
    config.header = {
      ...config.header,
      'Authorization': `Bearer ${token}`
    };
  }
  return config;
};

// 响应拦截器
const responseInterceptor = (response) => {
  // 统一处理响应
  if (response.statusCode === 401) {
    // 未授权，跳转到登录页
    uni.removeStorageSync('token');
    uni.removeStorageSync('userInfo');
    
    uni.showToast({
      title: '登录已过期，请重新登录',
      icon: 'none'
    });
    
    setTimeout(() => {
      uni.navigateTo({
        url: '/pages/login/login'
      });
    }, 1500);
  }
  
  return response;
};

// 封装请求方法
const request = (options) => {
  // 处理URL
  options.url = options.url.startsWith('http') ? options.url : BASE_URL + options.url;
  
  // 应用请求拦截器
  options = requestInterceptor(options);
  
  // 发起请求
  return new Promise((resolve, reject) => {
    uni.request({
      ...options,
      success: (res) => {
        // 应用响应拦截器
        const response = responseInterceptor(res);
        resolve(response);
      },
      fail: (err) => {
        reject(err);
      }
    });
  });
};

// 封装常用请求方法
const http = {
  get: (url, data = {}, options = {}) => {
    return request({
      url,
      data,
      method: 'GET',
      ...options
    });
  },
  post: (url, data = {}, options = {}) => {
    return request({
      url,
      data,
      method: 'POST',
      ...options
    });
  },
  put: (url, data = {}, options = {}) => {
    return request({
      url,
      data,
      method: 'PUT',
      ...options
    });
  },
  delete: (url, data = {}, options = {}) => {
    return request({
      url,
      data,
      method: 'DELETE',
      ...options
    });
  }
};

export default http;