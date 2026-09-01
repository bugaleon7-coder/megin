<template>
  <div
    id="userLayout"
    class="relative w-full h-full overflow-hidden bg-center bg-cover bg-no-repeat"
    :style="{ backgroundImage: `url(${loginBackground})` }"
  >
    <div
      class="absolute inset-0 bg-[linear-gradient(118deg,rgba(8,30,58,0.58),rgba(15,23,42,0.16)_52%,rgba(8,47,73,0.42))]"
    />
    <div class="relative z-10 flex min-h-full items-center justify-center px-5 py-16 md:min-h-screen">
      <div
        class="w-full max-w-md rounded-2xl border border-white/70 bg-white/68 px-7 py-10 shadow-[0_24px_70px_rgba(15,23,42,0.28)] backdrop-blur-2xl backdrop-saturate-150 transition-opacity duration-500 ease-out dark:border-slate-500/30 dark:bg-slate-900/68 sm:px-10 md:max-w-xl md:rounded-3xl md:px-14 md:py-12"
        :class="loginVisible ? 'opacity-100' : 'opacity-0 pointer-events-none'"
      >
        <div class="flex flex-col justify-between box-border">
          <div>
            <div class="mb-11">
              <p class="flex items-center justify-center gap-2.5 text-center text-4xl font-bold">
                <Logo :size="1.75" />
                <span>{{ $GIN_VUE_ADMIN.appName }}</span>
              </p>
              <p class="mt-3 text-center text-sm font-normal text-gray-500">
                初始帐号密码: admin/123456
              </p>
            </div>
            <el-form
              ref="loginForm"
              :model="loginFormData"
              :rules="rules"
              :validate-on-rule-change="false"
              @keyup.enter="submitForm"
            >
              <el-form-item prop="username" class="mb-6">
                <el-input
                  v-model="loginFormData.username"
                  size="large"
                  placeholder="请输入用户名"
                  suffix-icon="user"
                />
              </el-form-item>
              <el-form-item prop="password" class="mb-6">
                <el-input
                  v-model="loginFormData.password"
                  show-password
                  size="large"
                  type="password"
                  placeholder="请输入密码"
                />
              </el-form-item>
              <el-form-item v-if="totpEnabled" prop="otp" class="mb-6">
                <el-input
                  v-model="loginFormData.otp"
                  maxlength="6"
                  size="large"
                  :placeholder="`请输入6位Google验证码${totpIssuer ? `（${totpIssuer}）` : ''}`"
                />
              </el-form-item>
              <el-form-item
                v-if="loginFormData.openCaptcha"
                prop="captcha"
                class="mb-6"
              >
                <div class="flex w-full justify-between">
                  <el-input
                    v-model="loginFormData.captcha"
                    placeholder="请输入验证码"
                    size="large"
                    class="flex-1 mr-5"
                  />
                  <div class="w-1/3 h-11 bg-[#c3d4f2] rounded">
                    <img
                      v-if="picPath"
                      class="w-full h-full"
                      :src="picPath"
                      alt="请输入验证码"
                      @click="loadLoginConfig()"
                    />
                  </div>
                </div>
              </el-form-item>
              <el-form-item class="mb-6">
                <el-button
                  class="shadow shadow-active h-11 w-full"
                  type="primary"
                  size="large"
                  @click="submitForm"
                  >登 录</el-button
                >
              </el-form-item>
              <el-form-item v-if="isDev" class="mb-6">
                <el-button
                  class="shadow shadow-active h-11 w-full"
                  type="primary"
                  size="large"
                  @click="checkInit"
                  >前往初始化</el-button
                >
              </el-form-item>
            </el-form>
          </div>
        </div>
      </div>
    </div>

    <BottomInfo class="left-0 right-0 absolute bottom-3 mx-auto w-full z-20" />
  </div>
</template>

<script setup>
  import { captcha, getLoginConfig } from '@/api/user'
  import { checkDB } from '@/api/initdb'
  import BottomInfo from '@/components/bottomInfo/bottomInfo.vue'
  import { onMounted, reactive, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { useRouter } from 'vue-router'
  import { useUserStore } from '@/pinia/modules/user'
  import Logo from '@/components/logo/index.vue'
  import { isDev } from '@/utils/env.js'
  import loginBackground from '@/assets/login_right_banner.jpg'

  defineOptions({
    name: 'Login'
  })

  const backgroundReady = ref(false)
  const configReady = ref(false)
  const loginVisible = ref(false)

  const revealLogin = () => {
    if (!backgroundReady.value || !configReady.value || loginVisible.value) {
      return
    }
    requestAnimationFrame(() => {
      loginVisible.value = true
      window.setTimeout(() => document.getElementById('gva-loading-box')?.remove(), 80)
    })
  }

  const dismissInitialLoading = () => {
    backgroundReady.value = true
    revealLogin()
  }

  onMounted(() => {
    const image = new Image()
    image.onload = dismissInitialLoading
    image.onerror = dismissInitialLoading
    image.src = loginBackground
  })

  const router = useRouter()
  const captchaRequiredLength = ref(6)
  const totpEnabled = ref(false)
  const totpIssuer = ref('')
  // 验证函数
  const checkUsername = (rule, value, callback) => {
    if (value.length < 5) {
      return callback(new Error('请输入正确的用户名'))
    } else {
      callback()
    }
  }
  const checkPassword = (rule, value, callback) => {
    if (value.length < 6) {
      return callback(new Error('请输入正确的密码'))
    } else {
      callback()
    }
  }
  const checkCaptcha = (rule, value, callback) => {
    if (!loginFormData.openCaptcha) {
      return callback()
    }
    const sanitizedValue = (value || '').replace(/\s+/g, '')
    if (!sanitizedValue) {
      return callback(new Error('请输入验证码'))
    }
    if (!/^\d+$/.test(sanitizedValue)) {
      return callback(new Error('验证码须为数字'))
    }
    if (sanitizedValue.length < captchaRequiredLength.value) {
      return callback(
        new Error(`请输入至少${captchaRequiredLength.value}位数字验证码`)
      )
    }
    if (sanitizedValue !== value) {
      loginFormData.captcha = sanitizedValue
    }
    callback()
  }
  const checkOTP = (rule, value, callback) => {
    const sanitizedValue = (value || '').replace(/\s+/g, '')
    if (!sanitizedValue) {
      if (totpEnabled.value) {
        return callback(new Error('请输入Google验证码'))
      }
      return callback()
    }
    if (!/^\d{6}$/.test(sanitizedValue)) {
      return callback(new Error('Google验证码须为6位数字'))
    }
    if (sanitizedValue !== value) {
      loginFormData.otp = sanitizedValue
    }
    callback()
  }

  // 获取验证码
  const loadLoginConfig = async () => {
    try {
      const [captchaRes, configRes] = await Promise.all([captcha(), getLoginConfig()])
      captchaRequiredLength.value = Number(captchaRes.data?.captchaLength) || 0
      picPath.value = captchaRes.data?.picPath
      loginFormData.captchaId = captchaRes.data?.captchaId
      loginFormData.openCaptcha = captchaRes.data?.openCaptcha
      totpEnabled.value = Boolean(configRes.data?.totpEnabled)
      totpIssuer.value = configRes.data?.totpIssuer || ''
    } finally {
      configReady.value = true
      revealLogin()
    }
  }
  loadLoginConfig()

  // 登录相关操作
  const loginForm = ref(null)
  const picPath = ref('')
  const loginFormData = reactive({
    username: 'admin',
    password: '',
    otp: '',
    captcha: '',
    captchaId: '',
    openCaptcha: false
  })
  const rules = reactive({
    username: [{ validator: checkUsername, trigger: 'blur' }],
    password: [{ validator: checkPassword, trigger: 'blur' }],
    otp: [{ validator: checkOTP, trigger: 'blur' }],
    captcha: [{ validator: checkCaptcha, trigger: 'blur' }]
  })

  const userStore = useUserStore()
  const login = async () => {
    return await userStore.LoginIn(loginFormData)
  }
  const submitForm = () => {
    loginForm.value.validate(async (v) => {
      if (!v) {
        // 未通过前端静态验证
        ElMessage({
          type: 'error',
          message: '请正确填写登录信息',
          showClose: true
        })
        return false
      }

      // 通过验证，请求登陆
      const flag = await login()

      // 登陆失败，刷新验证码和登录配置
      if (!flag) {
        await loadLoginConfig()
        return false
      }

      // 登陆成功
      return true
    })
  }

  // 跳转初始化
  const checkInit = async () => {
    const res = await checkDB()
    if (res.code === 0) {
      if (res.data?.needInit) {
        userStore.NeedInit()
        await router.push({ name: 'Init' })
      } else {
        ElMessage({
          type: 'info',
          message: '已配置数据库信息，无法初始化'
        })
      }
    }
  }
</script>
