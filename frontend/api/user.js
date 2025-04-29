/**
 * 用户相关API
 */
import http from '@/utils/request';

export default {
  // 登录
  login(data) {
    return http.post('/auth/login', data);
  },
  
  // 注册
  register(data) {
    return http.post('/auth/register', data);
  },
  
  // 微信登录
  wechatLogin(code) {
    return http.post('/auth/wechat/login', { code });
  },
  
  // 获取用户信息
  getUserInfo() {
    return http.get('/api/user/info');
  },
  
  // 关注用户
  followUser(userId, action) {
    return http.post('/api/follow', { user_id: userId, action });
  },
  
  // 获取关注列表
  getFollowingList(page = 1, pageSize = 20) {
    return http.get('/api/follow/following', { page, page_size: pageSize });
  },
  
  // 获取粉丝列表
  getFollowersList(page = 1, pageSize = 20) {
    return http.get('/api/follow/followers', { page, page_size: pageSize });
  }
};