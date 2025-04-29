<script>
export default {
  onLaunch: function() {
    // 初始化云函数（如需）
    // uniCloud.init({ 
    //   provider: 'aliyun',
    //   spaceId: 'your-space-id'
    // })
    
    // 版本检测
    // const version = uni.getSystemInfoSync().SDKVersion 
    // if (compareVersion(version, '2.8.2') < 0) {
    //   uni.showModal({
    //     title: '警告',
    //     content: '当前基础库版本过低，请升级微信客户端'
    //   })
    // }
    
    // 添加全局导航守卫
    this.setupNavigationGuard();
  },
  methods: {
    // 设置全局导航守卫
    setupNavigationGuard() {
      // 需要登录才能访问的页面
      const authPages = [
        'pages/publish/publish',
        'pages/profile/profile'
      ];
      
      // 监听页面跳转
      uni.addInterceptor('navigateTo', {
        invoke: (args) => {
          // 检查是否需要登录
          if (this.needAuth(args.url, authPages)) {
            return this.checkLogin(args);
          }
          return args;
        }
      });
      
      uni.addInterceptor('switchTab', {
        invoke: (args) => {
          // 检查是否需要登录
          if (this.needAuth(args.url, authPages)) {
            return this.checkLogin(args);
          }
          return args;
        }
      });
    },
    
    // 检查页面是否需要登录
    needAuth(url, authPages) {
      return authPages.some(page => url.indexOf(page) !== -1);
    },
    
    // 检查登录状态
    checkLogin(args) {
      const token = uni.getStorageSync('token');
      if (!token) {
        uni.showToast({
          title: '请先登录',
          icon: 'none'
        });
        
        // 延迟跳转到登录页
        setTimeout(() => {
          uni.navigateTo({
            url: '/pages/login/login'
          });
        }, 1500);
        
        // 阻止原来的跳转
        return false;
      }
      return args;
    }
  }
}

// 版本比较方法
function compareVersion(v1, v2) {
  v1 = v1.split('.')
  v2 = v2.split('.')
  const len = Math.max(v1.length, v2.length)
  while (v1.length < len) v1.push('0')
  while (v2.length < len) v2.push('0')
  for (let i = 0; i < len; i++) {
    const num1 = parseInt(v1[i])
    const num2 = parseInt(v2[i])
    if (num1 > num2) return 1
    if (num1 < num2) return -1
  }
  return 0
}
</script>

<style lang="scss">
/* 全局样式 */
@import '@/uni_modules/uni-scss/index.scss';

/* 自定义主题变量 */
:root {
  --primary-color: #FF3366;  /* 小红书主色 */
  --secondary-color: #FF6B6B;
}

/* 跨平台样式 */
page {
  background-color: #f5f5f5;
  font-size: 28rpx;
  color: #333;
  padding-bottom: 100rpx !important;
}

/* 导航栏样式覆盖 */
.uni-nav-bar__header {
  box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.06) !important;
}

/* 按钮样式 */
uni-button[type=primary] {
  background-color: var(--primary-color) !important;
  border-radius: 48rpx !important;
  font-weight: 500 !important;
}

/* 图标兼容方案 */
.uni-icons {
  font-family: uniicons !important; /* 确保图标字体正确加载 */
}

/* 列表项统一样式 */
uni-list-item {
  border-bottom: 1rpx solid #eee !important;
  &::after {
    border: none !important;
  }
}

/* 输入框聚焦样式 */
uni-input:focus, 
uni-textarea:focus {
  border-color: var(--primary-color) !important;
  box-shadow: 0 0 8rpx rgba(255,51,102,0.1) !important;
}

/* 适配暗黑模式 */
@media (prefers-color-scheme: dark) {
  page {
    background-color: #121212;
    color: #fff;
  }
}
</style>