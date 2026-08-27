<template>
  <div class="profile-container">
    <!-- 顶部个人信息卡片 -->
    <div class="bg-white dark:bg-slate-800 rounded-2xl shadow-sm mb-8">
      <!-- 顶部背景图 -->
      <div class="h-48 bg-blue-50 dark:bg-slate-600 relative">
        <div class="absolute inset-0 bg-pattern opacity-7"></div>
      </div>

      <!-- 个人信息区 -->
      <div class="px-8 -mt-20 pb-8">
        <div class="flex flex-col lg:flex-row items-start gap-8">
          <!-- 左侧头像 -->
          <div class="profile-avatar-wrapper flex-shrink-0 mx-auto lg:mx-0">
            <SelectImage
                v-model="userStore.userInfo.headerImg"
                file-type="image"
                rounded
            />
          </div>

          <!-- 右侧信息 -->
          <div class="flex-1 pt-12 lg:pt-20 w-full">
            <div
              class="flex flex-col lg:flex-row items-start lg:items-start justify-between gap-4"
            >
              <div class="lg:mt-4">
                <div class="flex items-center gap-4 mb-4">
                  <div
                    v-if="!editFlag"
                    class="text-2xl font-bold flex items-center gap-3 text-gray-800 dark:text-gray-100"
                  >
                    {{ userStore.userInfo.nickName }}
                    <el-icon
                      class="cursor-pointer text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors duration-200"
                      @click="openEdit"
                    >
                      <edit />
                    </el-icon>
                  </div>
                  <div v-else class="flex items-center">
                    <el-input v-model="nickName" class="w-48 mr-4" />
                    <el-button type="primary" plain @click="enterEdit">
                      确认
                    </el-button>
                    <el-button type="danger" plain @click="closeEdit">
                      取消
                    </el-button>
                  </div>
                </div>

                <div
                  class="flex flex-col lg:flex-row items-start lg:items-center gap-4 lg:gap-8 text-gray-500 dark:text-gray-400"
                >
                  <div class="flex items-center gap-2">
                    <el-icon><location /></el-icon>
                    <span>中国·北京市·朝阳区</span>
                  </div>
                  <div class="flex items-center gap-2">
                    <el-icon><office-building /></el-icon>
                    <span>北京翻转极光科技有限公司</span>
                  </div>
                  <div class="flex items-center gap-2">
                    <el-icon><user /></el-icon>
                    <span>技术部·前端事业群</span>
                  </div>
                </div>
              </div>

              <div class="flex gap-4 mt-4">
                <el-button type="primary" plain icon="message">
                  发送消息
                </el-button>
                <el-button icon="share"> 分享主页 </el-button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 主要内容区 -->
    <div class="grid lg:grid-cols-12 md:grid-cols-1 gap-8">
      <!-- 左侧信息栏 -->
      <div class="lg:col-span-4">
        <div
          class="bg-white dark:bg-slate-800 rounded-xl p-6 mb-6 profile-card"
        >
          <h2 class="text-lg font-semibold mb-4 flex items-center gap-2">
            <el-icon class="text-blue-500"><info-filled /></el-icon>
            基本信息
          </h2>
          <div class="space-y-4">
            <div
              class="flex items-center gap-1 lg:gap-3 text-gray-600 dark:text-gray-300"
            >
              <el-icon class="text-blue-500"><phone /></el-icon>
              <span class="font-medium">手机号码：</span>
              <span>{{ userStore.userInfo.phone || '未设置' }}</span>
              <el-button
                link
                type="primary"
                class="ml-auto"
                @click="changePhoneFlag = true"
              >
                修改
              </el-button>
            </div>
            <div
              class="flex items-center gap-1 lg:gap-3 text-gray-600 dark:text-gray-300"
            >
              <el-icon class="text-green-500"><message /></el-icon>
              <span class="font-medium flex-shrink-0">邮箱地址：</span>
              <span>{{ userStore.userInfo.email || '未设置' }}</span>
              <el-button
                link
                type="primary"
                class="ml-auto"
                @click="changeEmailFlag = true"
              >
                修改
              </el-button>
            </div>
            <div
              class="flex items-center gap-1 lg:gap-3 text-gray-600 dark:text-gray-300"
            >
              <el-icon class="text-purple-500"><lock /></el-icon>
              <span class="font-medium">账号密码：</span>
              <span>已设置</span>
              <el-button
                link
                type="primary"
                class="ml-auto"
                @click="showPassword = true"
              >
                修改
              </el-button>
            </div>
            <div
              class="flex items-start gap-1 lg:gap-3 text-gray-600 dark:text-gray-300"
            >
              <el-icon class="text-amber-500 mt-1"><key /></el-icon>
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2">
                  <span class="font-medium">Google TOTP：</span>
                  <el-tag :type="totpStatus.enabled ? 'success' : 'info'" effect="light">
                    {{ totpStatus.enabled ? '已启用' : '未启用' }}
                  </el-tag>
                </div>
                <div class="text-sm text-gray-400 mt-1">
                  <span v-if="totpStatus.enabled && totpStatus.boundAt">
                    绑定时间：{{ formatDateTime(totpStatus.boundAt) }}
                  </span>
                  <span v-else-if="totpStatus.needSetup">
                    系统已开启双重验证，建议立即完成绑定。
                  </span>
                  <span v-else-if="!totpStatus.systemEnabled">
                    系统未开启 Google TOTP，绑定开关由后端控制。
                  </span>
                  <span v-else>
                    当前账号未启用动态码登录验证。
                  </span>
                </div>
              </div>
              <el-button
                v-if="!totpStatus.enabled && totpStatus.systemEnabled"
                link
                type="primary"
                class="ml-auto"
                @click="openTotpSetup"
              >
                {{ totpStatus.needSetup ? '立即绑定' : '绑定' }}
              </el-button>
              <el-button
                v-else-if="totpStatus.enabled"
                link
                type="danger"
                class="ml-auto"
                @click="disableTotpDialogVisible = true"
              >
                关闭
              </el-button>
            </div>
          </div>
        </div>

        <div class="bg-white dark:bg-slate-800 rounded-xl p-6 profile-card">
          <h2 class="text-lg font-semibold mb-4 flex items-center gap-2">
            <el-icon class="text-blue-500"><medal /></el-icon>
            技能特长
          </h2>
          <div class="flex flex-wrap gap-2">
            <el-tag effect="plain" type="success">GoLang</el-tag>
            <el-tag effect="plain" type="warning">JavaScript</el-tag>
            <el-tag effect="plain" type="danger">Vue</el-tag>
            <el-tag effect="plain" type="info">Gorm</el-tag>
            <el-button link class="text-sm">
              <el-icon><plus /></el-icon>
              添加技能
            </el-button>
          </div>
        </div>
      </div>

      <!-- 右侧内容区 -->
      <div class="lg:col-span-8">
        <div class="bg-white dark:bg-slate-800 rounded-xl p-6 profile-card">
          <el-tabs class="custom-tabs">
            <el-tab-pane>
              <template #label>
                <div class="flex items-center gap-2">
                  <el-icon><data-line /></el-icon>
                  数据统计
                </div>
              </template>
              <div class="grid grid-cols-2 md:grid-cols-4 gap-4 lg:gap-6 py-6">
                <div class="stat-card">
                  <div
                    class="text-2xl lg:text-4xl font-bold text-blue-500 mb-2"
                  >
                    138
                  </div>
                  <div class="text-gray-500 text-sm">项目参与</div>
                </div>
                <div class="stat-card">
                  <div
                    class="text-2xl lg:text-4xl font-bold text-green-500 mb-2"
                  >
                    2.3k
                  </div>
                  <div class="text-gray-500 text-sm">代码提交</div>
                </div>
                <div class="stat-card">
                  <div
                    class="text-2xl lg:text-4xl font-bold text-purple-500 mb-2"
                  >
                    95%
                  </div>
                  <div class="text-gray-500 text-sm">任务完成</div>
                </div>
                <div class="stat-card">
                  <div
                    class="text-2xl lg:text-4xl font-bold text-yellow-500 mb-2"
                  >
                    12
                  </div>
                  <div class="text-gray-500 text-sm">获得勋章</div>
                </div>
              </div>
            </el-tab-pane>
            <el-tab-pane>
              <template #label>
                <div class="flex items-center gap-2">
                  <el-icon><calendar /></el-icon>
                  近期动态
                </div>
              </template>
              <div class="py-6">
                <el-timeline>
                  <el-timeline-item
                    v-for="(activity, index) in activities"
                    :key="index"
                    :type="activity.type"
                    :timestamp="activity.timestamp"
                    :hollow="true"
                    class="pb-6"
                  >
                    <h3 class="text-base font-medium mb-1">
                      {{ activity.title }}
                    </h3>
                    <p class="text-gray-500 text-sm">{{ activity.content }}</p>
                  </el-timeline-item>
                </el-timeline>
              </div>
            </el-tab-pane>
          </el-tabs>
        </div>
      </div>
    </div>

    <!-- 弹窗 -->
    <el-dialog
      v-model="showPassword"
      title="修改密码"
      width="400px"
      class="custom-dialog"
      @close="clearPassword"
    >
      <el-form
        ref="modifyPwdForm"
        :model="pwdModify"
        :rules="rules"
        label-width="90px"
        class="py-4"
      >
        <el-form-item :minlength="6" label="原密码" prop="password">
          <el-input v-model="pwdModify.password" show-password />
        </el-form-item>
        <el-form-item :minlength="6" label="新密码" prop="newPassword">
          <el-input v-model="pwdModify.newPassword" show-password />
        </el-form-item>
        <el-form-item :minlength="6" label="确认密码" prop="confirmPassword">
          <el-input v-model="pwdModify.confirmPassword" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showPassword = false">取 消</el-button>
          <el-button type="primary" @click="savePassword">确 定</el-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog
      v-model="changePhoneFlag"
      title="修改手机号"
      width="400px"
      class="custom-dialog"
    >
      <el-form :model="phoneForm" label-width="80px" class="py-4">
        <el-form-item label="手机号">
          <el-input v-model="phoneForm.phone" placeholder="请输入新的手机号码">
            <template #prefix>
              <el-icon><phone /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="验证码">
          <div class="flex gap-4">
            <el-input
              v-model="phoneForm.code"
              placeholder="请输入验证码[模拟]"
              class="flex-1"
            >
              <template #prefix>
                <el-icon><key /></el-icon>
              </template>
            </el-input>
            <el-button
              type="primary"
              :disabled="time > 0"
              class="w-32"
              @click="getCode"
            >
              {{ time > 0 ? `${time}s` : '获取验证码' }}
            </el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="closeChangePhone">取 消</el-button>
          <el-button type="primary" @click="changePhone">确 定</el-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog
      v-model="changeEmailFlag"
      title="修改邮箱"
      width="400px"
      class="custom-dialog"
    >
      <el-form :model="emailForm" label-width="80px" class="py-4">
        <el-form-item label="邮箱">
          <el-input v-model="emailForm.email" placeholder="请输入新的邮箱地址">
            <template #prefix>
              <el-icon><message /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="验证码">
          <div class="flex gap-4">
            <el-input
              v-model="emailForm.code"
              placeholder="请输入验证码[模拟]"
              class="flex-1"
            >
              <template #prefix>
                <el-icon><key /></el-icon>
              </template>
            </el-input>
            <el-button
              type="primary"
              :disabled="emailTime > 0"
              class="w-32"
              @click="getEmailCode"
            >
              {{ emailTime > 0 ? `${emailTime}s` : '获取验证码' }}
            </el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="closeChangeEmail">取 消</el-button>
          <el-button type="primary" @click="changeEmail">确 定</el-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog
      v-model="totpDialogVisible"
      title="绑定 Google TOTP"
      width="420px"
      class="custom-dialog"
      @close="resetTotpSetup"
    >
      <div class="py-2" v-loading="totpSetupLoading">
        <div class="text-sm text-gray-500 mb-4">
          使用 Google Authenticator、Microsoft Authenticator 等应用扫码，随后输入 6 位动态码确认启用。
        </div>
        <div
          v-if="totpSetup.qrContent"
          class="flex justify-center rounded-lg bg-gray-50 py-4 mb-4"
        >
          <vue-qr
            :text="totpSetup.qrContent"
            :size="220"
            :margin="0"
            color-dark="#111827"
            color-light="#ffffff"
          />
        </div>
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="签发方">
            {{ totpStatus.issuer || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="账号">
            {{ totpStatus.account || userStore.userInfo.userName || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="密钥">
            <div class="break-all">{{ totpSetup.secret || '-' }}</div>
          </el-descriptions-item>
        </el-descriptions>
        <el-form class="mt-4" :model="totpEnableForm" label-width="88px">
          <el-form-item label="动态码">
            <el-input
              v-model="totpEnableForm.otp"
              maxlength="6"
              placeholder="请输入 6 位动态码"
            />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="totpDialogVisible = false">取 消</el-button>
          <el-button type="primary" :loading="totpSubmitting" @click="confirmEnableTotp">
            确认启用
          </el-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog
      v-model="disableTotpDialogVisible"
      title="关闭 Google TOTP"
      width="400px"
      class="custom-dialog"
      @close="disableTotpForm.otp = ''"
    >
      <div class="text-sm text-gray-500 mb-4">
        关闭后登录将不再校验动态码，请输入当前 6 位 Google 验证码确认。
      </div>
      <el-form :model="disableTotpForm" label-width="88px">
        <el-form-item label="动态码">
          <el-input
            v-model="disableTotpForm.otp"
            maxlength="6"
            placeholder="请输入 6 位动态码"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="disableTotpDialogVisible = false">取 消</el-button>
          <el-button type="danger" :loading="totpSubmitting" @click="confirmDisableTotp">
            确认关闭
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import {
    setSelfInfo,
    changePassword,
    getTotpStatus,
    initTotp,
    enableTotp,
    disableTotp
  } from '@/api/user.js'
  import vueQr from 'vue-qr/src/packages/vue-qr.vue'
  import { onMounted, reactive, ref, watch } from 'vue'
  import { ElMessage } from 'element-plus'
  import { useUserStore } from '@/pinia/modules/user'
  import SelectImage from '@/components/selectImage/selectImage.vue'
  defineOptions({
    name: 'Person'
  })

  const userStore = useUserStore()
  const modifyPwdForm = ref(null)
  const showPassword = ref(false)
  const pwdModify = ref({})
  const nickName = ref('')
  const editFlag = ref(false)
  const totpStatus = reactive({
    enabled: false,
    boundAt: '',
    issuer: '',
    account: '',
    needSetup: false,
    systemEnabled: false
  })
  const totpDialogVisible = ref(false)
  const disableTotpDialogVisible = ref(false)
  const totpSetupLoading = ref(false)
  const totpSubmitting = ref(false)
  const totpSetup = reactive({
    secret: '',
    qrContent: '',
    otpAuthUrl: ''
  })
  const totpEnableForm = reactive({
    otp: ''
  })
  const disableTotpForm = reactive({
    otp: ''
  })

  const rules = reactive({
    password: [
      { required: true, message: '请输入密码', trigger: 'blur' },
      { min: 6, message: '最少6个字符', trigger: 'blur' }
    ],
    newPassword: [
      { required: true, message: '请输入新密码', trigger: 'blur' },
      { min: 6, message: '最少6个字符', trigger: 'blur' }
    ],
    confirmPassword: [
      { required: true, message: '请输入确认密码', trigger: 'blur' },
      { min: 6, message: '最少6个字符', trigger: 'blur' },
      {
        validator: (rule, value, callback) => {
          if (value !== pwdModify.value.newPassword) {
            callback(new Error('两次密码不一致'))
          } else {
            callback()
          }
        },
        trigger: 'blur'
      }
    ]
  })

  const savePassword = async () => {
    modifyPwdForm.value.validate((valid) => {
      if (valid) {
        changePassword({
          password: pwdModify.value.password,
          newPassword: pwdModify.value.newPassword
        }).then((res) => {
          if (res.code === 0) {
            ElMessage.success('修改密码成功！')
          }
          showPassword.value = false
        })
      }
    })
  }

  const clearPassword = () => {
    pwdModify.value = {
      password: '',
      newPassword: '',
      confirmPassword: ''
    }
    modifyPwdForm.value?.clearValidate()
  }

  const openEdit = () => {
    nickName.value = userStore.userInfo.nickName
    editFlag.value = true
  }

  const closeEdit = () => {
    nickName.value = ''
    editFlag.value = false
  }

  const enterEdit = async () => {
    const res = await setSelfInfo({
      nickName: nickName.value
    })
    if (res.code === 0) {
      userStore.ResetUserInfo({ nickName: nickName.value })
      ElMessage.success('修改成功')
    }
    nickName.value = ''
    editFlag.value = false
  }

  const changePhoneFlag = ref(false)
  const time = ref(0)
  const phoneForm = reactive({
    phone: '',
    code: ''
  })

  const getCode = async () => {
    time.value = 60
    let timer = setInterval(() => {
      time.value--
      if (time.value <= 0) {
        clearInterval(timer)
        timer = null
      }
    }, 1000)
  }

  const closeChangePhone = () => {
    changePhoneFlag.value = false
    phoneForm.phone = ''
    phoneForm.code = ''
  }

  const changePhone = async () => {
    const res = await setSelfInfo({ phone: phoneForm.phone })
    if (res.code === 0) {
      ElMessage.success('修改成功')
      userStore.ResetUserInfo({ phone: phoneForm.phone })
      closeChangePhone()
    }
  }

  const changeEmailFlag = ref(false)
  const emailTime = ref(0)
  const emailForm = reactive({
    email: '',
    code: ''
  })

  const getEmailCode = async () => {
    emailTime.value = 60
    let timer = setInterval(() => {
      emailTime.value--
      if (emailTime.value <= 0) {
        clearInterval(timer)
        timer = null
      }
    }, 1000)
  }

  const closeChangeEmail = () => {
    changeEmailFlag.value = false
    emailForm.email = ''
    emailForm.code = ''
  }

  const changeEmail = async () => {
    const res = await setSelfInfo({ email: emailForm.email })
    if (res.code === 0) {
      ElMessage.success('修改成功')
      userStore.ResetUserInfo({ email: emailForm.email })
      closeChangeEmail()
    }
  }

  const formatDateTime = (value) => {
    if (!value) {
      return '-'
    }
    const date = new Date(value)
    if (Number.isNaN(date.getTime())) {
      return value
    }
    const pad = (num) => String(num).padStart(2, '0')
    return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
  }

  const syncTotpStatus = (data = {}) => {
    totpStatus.enabled = Boolean(data.enabled)
    totpStatus.boundAt = data.boundAt || ''
    totpStatus.issuer = data.issuer || ''
    totpStatus.account = data.account || ''
    totpStatus.needSetup = Boolean(data.needSetup)
    totpStatus.systemEnabled = Boolean(data.systemEnabled)
  }

  const fetchTotpStatus = async () => {
    const res = await getTotpStatus()
    if (res.code === 0) {
      syncTotpStatus(res.data || {})
    }
  }

  const resetTotpSetup = () => {
    totpEnableForm.otp = ''
    totpSetup.secret = ''
    totpSetup.qrContent = ''
    totpSetup.otpAuthUrl = ''
  }

  const openTotpSetup = async () => {
    if (!totpStatus.systemEnabled) {
      ElMessage.warning('系统未开启 Google TOTP')
      return
    }
    totpSetupLoading.value = true
    try {
      const res = await initTotp()
      if (res.code === 0) {
        totpSetup.secret = res.data?.secret || ''
        totpSetup.qrContent = res.data?.qrContent || res.data?.otpAuthUrl || ''
        totpSetup.otpAuthUrl = res.data?.otpAuthUrl || ''
        totpDialogVisible.value = true
      }
    } finally {
      totpSetupLoading.value = false
    }
  }

  const normalizeOtp = (value) => (value || '').replace(/\s+/g, '')

  const confirmEnableTotp = async () => {
    const otp = normalizeOtp(totpEnableForm.otp)
    if (!/^\d{6}$/.test(otp)) {
      ElMessage.error('请输入 6 位数字动态码')
      return
    }
    totpSubmitting.value = true
    try {
      const res = await enableTotp({ otp })
      if (res.code === 0) {
        ElMessage.success('Google TOTP 已启用')
        totpDialogVisible.value = false
        resetTotpSetup()
        await fetchTotpStatus()
      }
    } finally {
      totpSubmitting.value = false
    }
  }

  const confirmDisableTotp = async () => {
    const otp = normalizeOtp(disableTotpForm.otp)
    if (!/^\d{6}$/.test(otp)) {
      ElMessage.error('请输入 6 位数字动态码')
      return
    }
    totpSubmitting.value = true
    try {
      const res = await disableTotp({ otp })
      if (res.code === 0) {
        ElMessage.success('Google TOTP 已关闭')
        disableTotpDialogVisible.value = false
        disableTotpForm.otp = ''
        await fetchTotpStatus()
      }
    } finally {
      totpSubmitting.value = false
    }
  }

  watch(() => userStore.userInfo.headerImg, async(val) => {
    const res = await setSelfInfo({ headerImg: val })
    if (res.code === 0) {
      userStore.ResetUserInfo({ headerImg: val })
      ElMessage({
        type: 'success',
        message: '设置成功',
      })
    }
  })

  onMounted(() => {
    fetchTotpStatus()
  })

  // 添加活动数据
  const activities = [
    {
      timestamp: '2024-01-10',
      title: '完成项目里程碑',
      content: '成功完成第三季度主要项目开发任务，获得团队一致好评',
      type: 'primary'
    },
    {
      timestamp: '2024-01-11',
      title: '代码审核完成',
      content: '完成核心模块代码审核，提出多项改进建议并获采纳',
      type: 'success'
    },
    {
      timestamp: '2024-01-12',
      title: '技术分享会',
      content: '主持团队技术分享会，分享前端性能优化经验',
      type: 'warning'
    },
    {
      timestamp: '2024-01-13',
      title: '新功能上线',
      content: '成功上线用户反馈的新特性，显著提升用户体验',
      type: 'danger'
    }
  ]
</script>

<style lang="scss">
  .profile-container {
    @apply p-4 lg:p-6 min-h-screen bg-gray-50 dark:bg-slate-900;

    .bg-pattern {
      background-image: url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23000000' fill-opacity='0.1'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E");
    }

    .profile-card {
      @apply shadow-sm hover:shadow-md transition-shadow duration-300;
    }

    .profile-action-btn {
      @apply bg-white/10 hover:bg-white/20 border-white/20;
      .el-icon {
        @apply mr-1;
      }
    }

    .stat-card {
      @apply p-4 lg:p-6 rounded-lg bg-gray-50 dark:bg-slate-700/50 text-center hover:shadow-md transition-all duration-300;
    }

    .custom-tabs {
      :deep(.el-tabs__nav-wrap::after) {
        @apply h-0.5 bg-gray-100 dark:bg-gray-700;
      }
      :deep(.el-tabs__active-bar) {
        @apply h-0.5 bg-blue-500;
      }
      :deep(.el-tabs__item) {
        @apply text-base font-medium px-6;
        .el-icon {
          @apply mr-1 text-lg;
        }
        &.is-active {
          @apply text-blue-500;
        }
      }
      :deep(.el-timeline-item__node--normal) {
        @apply left-[-2px];
      }
      :deep(.el-timeline-item__wrapper) {
        @apply pl-8;
      }
      :deep(.el-timeline-item__timestamp) {
        @apply text-gray-400 text-sm;
      }
    }

    .custom-dialog {
      :deep(.el-dialog__header) {
        @apply mb-0 pb-4 border-b border-gray-100 dark:border-gray-700;
      }
      :deep(.el-dialog__footer) {
        @apply mt-0 pt-4 border-t border-gray-100 dark:border-gray-700;
      }
      :deep(.el-input__wrapper) {
        @apply shadow-none;
      }
      :deep(.el-input__prefix) {
        @apply mr-2;
      }
    }

    .edit-input {
      :deep(.el-input__wrapper) {
        @apply bg-white/10 border-white/20 shadow-none;
        input {
          @apply text-white;
          &::placeholder {
            @apply text-white/60;
          }
        }
      }
    }
  }
</style>
