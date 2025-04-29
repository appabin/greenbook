<template>
  <view class="register-container">
    <view class="header">
      <text class="title">注册账号</text>
      <text class="subtitle">创建一个新账号，开始探索绿书</text>
    </view>
    
    <view class="form-box">
      <view class="input-group">
        <text class="label">用户名</text>
        <input class="input" v-model="registerForm.username" placeholder="请输入用户名" />
      </view>
      
      <view class="input-group">
        <text class="label">昵称</text>
        <input class="input" v-model="registerForm.nickname" placeholder="请输入昵称" />
      </view>
      
      <view class="input-group">
        <text class="label">手机号</text>
        <input class="input" v-model="registerForm.phone" type="number" placeholder="请输入手机号" maxlength="11" />
      </view>
      
      <view class="input-group">
        <text class="label">邮箱</text>
        <input class="input" v-model="registerForm.email" type="email" placeholder="请输入邮箱" />
      </view>
      
      <view class="input-group">
        <text class="label">密码</text>
        <input class="input" v-model="registerForm.password" type="password" placeholder="请输入密码" />
      </view>
      
      <view class="input-group">
        <text class="label">确认密码</text>
        <input class="input" v-model="confirmPassword" type="password" placeholder="请再次输入密码" />
      </view>
      
      <button class="submit-btn" type="primary" @click="handleRegister">注册</button>
      
      <view class="login-link">
        已有账号? <text class="link" @click="goToLogin">去登录</text>
      </view>
    </view>
  </view>
</template>

<script>
export default {
  data() {
    return {
      registerForm: {
        username: '',
        nickname: '',
        phone: '',
        email: '',
        password: ''
      },
      confirmPassword: ''
    }
  },
  methods: {
    // 处理注册
    handleRegister() {
      // 表单验证
      if (!this.registerForm.username || !this.registerForm.password) {
        uni.showToast({
          title: '用户名和密码不能为空',
          icon: 'none'
        })
        return
      }
      
      if (!this.registerForm.nickname) {
        uni.showToast({
          title: '昵称不能为空',
          icon: 'none'
        })
        return
      }
      
      if (!this.registerForm.email) {
        uni.showToast({
          title: '邮箱不能为空',
          icon: 'none'
        })
        return
      }
      
      if (!this.registerForm.phone) {
        uni.showToast({
          title: '手机号不能为空',
          icon: 'none'
        })
        return
      }
      
      if (this.registerForm.password !== this.confirmPassword) {
        uni.showToast({
          title: '两次输入的密码不一致',
          icon: 'none'
        })
        return
      }
      
      // 调用注册接口
      uni.request({
        url: 'http://localhost:8080/auth/register',
        method: 'POST',
        data: this.registerForm,
        success: (res) => {
          if (res.statusCode === 200) {
            // 保存token和用户信息
            uni.setStorageSync('token', res.data.token)
            uni.setStorageSync('userInfo', res.data.user)
            
            uni.showToast({
              title: '注册成功'
            })
            
            // 跳转到首页
            setTimeout(() => {
              uni.switchTab({
                url: '/pages/recommend/recommend'
              })
            }, 1500)
          } else {
            uni.showToast({
              title: res.data.error || '注册失败',
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
    
    // 跳转到登录页面
    goToLogin() {
      uni.navigateTo({
        url: '/pages/login/login'
      })
    }
  }
}
</script>

<style lang="scss">
.register-container {
  padding: 40rpx 50rpx;
  
  .header {
    margin-bottom: 60rpx;
    
    .title {
      font-size: 48rpx;
      font-weight: bold;
      color: #333;
      margin-bottom: 20rpx;
    }
    
    .subtitle {
      font-size: 28rpx;
      color: #999;
    }
  }
  
  .form-box {
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
    
    .login-link {
      text-align: center;
      margin-top: 30rpx;
      font-size: 28rpx;
      color: #666;
      
      .link {
        color: var(--primary-color);
      }
    }
  }
}
</style>