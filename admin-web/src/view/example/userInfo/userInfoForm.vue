
<template>
  <div>
    <div class="gva-form-box">
      <el-form :model="formData" ref="elFormRef" label-position="right" :rules="rule" label-width="80px">
        <el-form-item label="ID:" prop="id">
    <el-input v-model.number="formData.id" :clearable="true" placeholder="请输入ID" />
</el-form-item>
        <el-form-item label="登录名称:" prop="login_name">
    <el-input v-model="formData.login_name" :clearable="true" placeholder="请输入登录名称" />
</el-form-item>
        <el-form-item label="密码:" prop="password">
    <el-input v-model="formData.password" :clearable="true" placeholder="请输入密码" />
</el-form-item>
        <el-form-item label="createdAt字段:" prop="created_at">
    <el-date-picker v-model="formData.created_at" type="date" style="width:100%" placeholder="选择日期" :clearable="true" />
</el-form-item>
        <el-form-item label="updatedAt字段:" prop="updated_at">
    <el-date-picker v-model="formData.updated_at" type="date" style="width:100%" placeholder="选择日期" :clearable="true" />
</el-form-item>
        <el-form-item label="salt字段:" prop="salt">
    <el-input v-model="formData.salt" :clearable="true" placeholder="请输入salt字段" />
</el-form-item>
        <el-form-item label="最后登录时间:" prop="last_login_time">
    <el-date-picker v-model="formData.last_login_time" type="date" style="width:100%" placeholder="选择日期" :clearable="true" />
</el-form-item>
        <el-form-item label="手动验证码:" prop="mobile_authcode">
    <el-input v-model="formData.mobile_authcode" :clearable="true" placeholder="请输入手动验证码" />
</el-form-item>
        <el-form-item label="token字段:" prop="token">
    <el-input v-model="formData.token" :clearable="true" placeholder="请输入token字段" />
</el-form-item>
        <el-form-item label="手机号:" prop="mobile">
    <el-input v-model="formData.mobile" :clearable="true" placeholder="请输入手机号" />
</el-form-item>
        <el-form-item>
          <el-button :loading="btnLoading" type="primary" @click="save">保存</el-button>
          <el-button type="primary" @click="back">返回</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import {
  createUserInfo,
  updateUserInfo,
  findUserInfo
} from '@/api/example/userInfo'

defineOptions({
    name: 'UserInfoForm'
})

// 自动获取字典
import { getDictFunc } from '@/utils/format'
import { useRoute, useRouter } from "vue-router"
import { ElMessage } from 'element-plus'
import { ref, reactive } from 'vue'


const route = useRoute()
const router = useRouter()

// 提交按钮loading
const btnLoading = ref(false)

const type = ref('')
const formData = ref({
            id: undefined,
            login_name: '',
            password: '',
            created_at: new Date(),
            updated_at: new Date(),
            salt: '',
            last_login_time: new Date(),
            mobile_authcode: '',
            token: '',
            mobile: '',
        })
// 验证规则
const rule = reactive({
})

const elFormRef = ref()

// 初始化方法
const init = async () => {
 // 建议通过url传参获取目标数据ID 调用 find方法进行查询数据操作 从而决定本页面是create还是update 以下为id作为url参数示例
    if (route.query.id) {
      const res = await findUserInfo({ ID: route.query.id })
      if (res.code === 0) {
        formData.value = res.data
        type.value = 'update'
      }
    } else {
      type.value = 'create'
    }
}

init()
// 保存按钮
const save = async() => {
      btnLoading.value = true
      elFormRef.value?.validate( async (valid) => {
         if (!valid) return btnLoading.value = false
            let res
           switch (type.value) {
             case 'create':
               res = await createUserInfo(formData.value)
               break
             case 'update':
               res = await updateUserInfo(formData.value)
               break
             default:
               res = await createUserInfo(formData.value)
               break
           }
           btnLoading.value = false
           if (res.code === 0) {
             ElMessage({
               type: 'success',
               message: '创建/更改成功'
             })
           }
       })
}

// 返回按钮
const back = () => {
    router.go(-1)
}

</script>

<style>
</style>
