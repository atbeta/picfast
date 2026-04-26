export const resources = {
  'zh-CN': {
    translation: {
      appName: 'PicFast',
      nav: {
        homeUpload: '游客上传',
        login: '登录',
        register: '注册',
        upload: '上传图片',
        images: '图片',
        albums: '相册',
        apiTokens: '令牌',
        settings: '设置',
        admin: '管理员',
      },
      common: {
        language: '语言',
        theme: '主题',
        logout: '退出登录',
      },
      page: {
        guestUpload: {
          title: '游客上传',
          subtitle: '无需登录也可上传（受站点开关控制）',
        },
        login: { title: '登录' },
        register: { title: '注册' },
        upload: { title: '上传图片' },
        images: { title: '图片管理' },
        albums: { title: '相册管理' },
        apiTokens: { title: 'API 令牌' },
        settings: { title: '个人设置' },
      },
    },
  },
  'en-US': {
    translation: {
      appName: 'PicFast',
      nav: {
        homeUpload: 'Guest Upload',
        login: 'Login',
        register: 'Register',
        upload: 'Upload',
        images: 'Images',
        albums: 'Albums',
        apiTokens: 'Tokens',
        settings: 'Settings',
        admin: 'Admin',
      },
      common: {
        language: 'Language',
        theme: 'Theme',
        logout: 'Logout',
      },
      page: {
        guestUpload: {
          title: 'Guest Upload',
          subtitle: 'Upload without login (controlled by site settings)',
        },
        login: { title: 'Login' },
        register: { title: 'Register' },
        upload: { title: 'Upload Images' },
        images: { title: 'Image Management' },
        albums: { title: 'Album Management' },
        apiTokens: { title: 'API Tokens' },
        settings: { title: 'Profile Settings' },
      },
    },
  },
} as const

export const defaultLng = 'zh-CN'
export const supportedLngs = ['zh-CN', 'en-US'] as const
export type SupportedLng = (typeof supportedLngs)[number]
