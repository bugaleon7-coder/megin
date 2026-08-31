/*
 * MeAdmin 前端初始化
 *
 * */
// 加载网站配置文件夹
import { register } from './global'
import packageInfo from '../../package.json'

export default {
  install: (app) => {
    register(app)
    console.log(`欢迎使用 MeAdmin\n当前版本:v${packageInfo.version}`)
  }
}
