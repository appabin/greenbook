<template>
  <view class="login-container">
    <view class="logo-box">
      <image class="logo" src="/static/logo.png" mode="aspectFit"></image>
      <text class="title">绿书</text>
    </view>
    
    <view class="form-box">
      <view class="input-group">
        <text class="label">用户名</text>
        <input class="input" v-model="loginForm.username" placeholder="请输入用户名" />
      </view>
      <view class="input-group">
        <text class="label">密码</text>
        <input class="input" v-model="loginForm.password" type="password" placeholder="请输入密码" />
      </view>
      
      <button class="submit-btn" type="primary" @click="handleLogin">登录</button>
      
      <view class="options">
        <text class="link" @click="goToRegister">注册账号</text>
        <text class="link">忘记密码?</text>
      </view>
    </view>
    
    <view class="other-login">
      <view class="divider">
        <text class="divider-text">其他登录方式</text>
      </view>
      <view class="icon-group">
        <view class="icon-item" @click="handleWechatLogin">
          <text class="iconfont icon-wechat"></text>
          <text class="icon-text">微信</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
export default {
  data() {
    return {
      loginForm: {
        username: '',
        password: ''
      }
    }
  },
  methods: {
    // 处理登录
    handleLogin() {
      if (!this.loginForm.username || !this.loginForm.password) {
        uni.showToast({
          title: '用户名和密码不能为空',
          icon: 'none'
        })
        return
      }
      
      // 调用登录接口
      uni.request({
        url: 'http://localhost:8080/auth/login',
        method: 'POST',
        data: this.loginForm,
        success: (res) => {
          if (res.statusCode === 200) {
            // 保存token和用户信息
            uni.setStorageSync('token', res.data.token)
            uni.setStorageSync('userInfo', res.data.user)
            
            // 跳转到首页
            uni.switchTab({
              url: '/pages/recommend/recommend'
            })
            
            uni.showToast({
              title: '登录成功'
            })
          } else {
            uni.showToast({
              title: res.data.error || '登录失败',
              icon: 'none'
            })
          }
        },
        fail: () => {
          uni.showToast({
            title: '网络错误，请稍后再试',
            icon: 'none'
          })
        }
      })
    },
    
    // 跳转到注册页面
    goToRegister() {
      uni.navigateTo({
        url: '/pages/register/register'
      })
    },
    
    // 微信登录
    handleWechatLogin() {
      // 仅在微信小程序环境下可用
      // #ifdef MP-WEIXIN
      uni.login({
        provider: 'weixin',
        success: (loginRes) => {
          if (loginRes.code) {
            // 获取到微信登录code后调用后端接口
            uni.request({
              url: 'http://localhost:8080/auth/wechat/login',
              method: 'POST',
              data: {
                code: loginRes.code
              },
              success: (res) => {
                if (res.statusCode === 200) {
                  uni.setStorageSync('token', res.data.token)
                  uni.setStorageSync('userInfo', res.data.user)
                  
                  uni.switchTab({
                    url: '/pages/recommend/recommend'
                  })
                } else {
                  uni.showToast({
                    title: res.data.error || '微信登录失败',
                    icon: 'none'
                  })
                }
              }
            })
          }
        }
      })
      // #endif
      
      // 非微信环境提示
      // #ifndef MP-WEIXIN
      uni.showToast({
        title: '请在微信小程序中使用微信登录',
        icon: 'none'
      })
      // #endif
    }
  }
}
</script>

<style lang="scss">
.login-container {
  padding: 60rpx 50rpx;
  
  .logo-box {
    display: flex;
    flex-direction: column;
    align-items: center;
    margin-bottom: 80rpx;
    
    .logo {
      width: 160rpx;
      height: 160rpx;
      margin-bottom: 20rpx;
    }
    
    .title {
      font-size: 48rpx;
      font-weight: bold;
      color: #333;
    }
  }
  
  .form-box {
    margin-bottom: 60rpx;
    
    .input-group {
      margin-bottom: 30rpx;
      
      .label {
        display: block;
        margin-bottom: 10rpx;
        font-size: 28rpx;
        color: #666;
      }
      
      .input {
        width: 100%;
        height: 90rpx;
        background-color: #f5f5f5;
        border-radius: 45rpx;
        padding: 0 30rpx;
        font-size: 30rpx;
      }
    }
    
    .submit-btn {
      width: 100%;
      height: 90rpx;
      line-height: 90rpx;
      margin-top: 50rpx;
      border-radius: 45rpx;
      font-size: 32rpx;
    }
    
    .options {
      display: flex;
      justify-content: space-between;
      margin-top: 30rpx;
      
      .link {
        font-size: 28rpx;
        color: #666;
      }
    }
  }
  
  .other-login {
    .divider {
      position: relative;
      text-align: center;
      margin: 40rpx 0;
      
      &::before {
        content: '';
        position: absolute;
        top: 50%;
        left: 0;
        width: 100%;
        height: 1px;
        background-color: #eee;
        transform: translateY(-50%);
      }
      
      .divider-text {
        position: relative;
        display: inline-block;
        padding: 0 20rpx;
        background-color: #fff;
        font-size: 28rpx;
        color: #999;
      }
    }
    
    .icon-group {
      display: flex;
      justify-content: center;
      
      .icon-item {
        display: flex;
        flex-direction: column;
        align-items: center;
        margin: 0 40rpx;
        
        .iconfont {
          font-size: 80rpx;
          color: #07c160;
        }
        
        .icon-text {
          margin-top: 10rpx;
          font-size: 24rpx;
          color: #666;
        }
      }
    }
  }
}
</style>