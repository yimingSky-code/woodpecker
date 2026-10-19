# 手动运行与 Deploy 参数

本 fork 基于 Woodpecker `v3.19.0`，参考 [PR #6878](https://github.com/woodpecker-ci/woodpecker/pull/6878) 的仓库参数模型、CRUD 与设置页，并补齐 Deploy、多选及实际输入持久化。该 PR 尚未合并；本功能属于本 fork，不代表官方发布版已经支持。

## 页面操作

在仓库 **Settings → Parameters** 中定义参数。设置页沿用仓库管理员入口；参数 API 与原仓库 Secrets、Registry API 一样需要 push 权限。每个参数选择运行入口：

- **Manual run**：出现在原生手动运行页，值传入本次 Pipeline 的 variables。
- **Deploy from a build**：出现在成功构建记录的原生 Deploy 弹窗，使用该记录的来源和配置。不会新建 manual 运行再取分支 HEAD。

同一仓库的同名参数可以在两个入口独立配置，各自的必填规则互不影响。未配置参数的仓库继续使用原来的自由变量输入；push、pull request、cron 和普通 Restart 不应用本次表单规则。

| 类型 | 控件 | 传给 workflow 的值 |
| --- | --- | --- |
| `string` | 文本框 | 字符串 |
| `number` | 数字输入 | 有限数字的字符串 |
| `boolean` | Checkbox | `true` 或 `false` |
| `choice` | 下拉单选 | 一个配置选项 |
| `multi_choice` | Checkbox 多选组 | 按配置顺序以逗号连接，如 `server,worker` |

选项为仓库级静态列表。当前没有按分支、OPERATION、环境或 Artifact 联动的选项加载，也没有动态数据库备份列表。选项不是发布权限、确认或执行器的业务校验。

## 配置示例

以下是表单定义的 API 示例，可直接在 Settings 页面逐项创建；服务名称与环境值必须使用实际受控 workflow / Go 接口接受的值。

```json
{
  "event": "manual",
  "name": "BUILD_SERVICES",
  "type": "multi_choice",
  "options": ["server", "scheduler", "worker"],
  "default": "server,worker",
  "required": true,
  "description": "选择本次构建的服务"
}
```

用户可同时勾选 server 和 worker，不必手写服务列表。多选值继续使用原有变量通道，不增加独立 Plan 或参数传输服务。多选选项不能包含逗号，重复或未知值会被 Server 拒绝。

```json
{
  "event": "deployment",
  "name": "deploy_to",
  "type": "choice",
  "options": ["testing", "production"],
  "required": true,
  "description": "选择发布环境"
}
```

`deployment` 的 `deploy_to` 和 `deploy_task` 对应原生 Deploy 的目标环境与任务字段，只允许 string / choice；它们不是普通环境变量。可将 `deploy_task` 配成 `hot` / `cold` 下拉。未定义时仍使用原文本输入。其他 deployment 参数进入 variables，例如服务多选或原有具名确认输入。

构建本身不要求环境。自动发布可将 `AUTO_DEPLOY` 定义为 manual boolean，将 `AUTO_DEPLOY_TARGET` 定义为 manual choice；它们仍由受控 workflow 消费，不会自动成为原生 Deploy 目标。

## 默认值、继承与校验

- 缺失字段才应用默认值。显式取消所有服务勾选得到空值：必填时提示并拒绝；可选时保留空值，不能重新填回默认服务。
- Deploy 从父记录预填同名变量、目标和任务；新输入覆盖继承值，缺失字段使用继承值或定义默认值。
- 新部署保存实际生效的 variables、deploy_to、deploy_task。父构建的变量和目标不变。再次从部署记录 Deploy 时可继承上次实际参数。
- 原生 Parent、commit、分支和配置继承保持。表单不更换镜像、不重建、不证明 Artifact 存在；workflow / Go 仍负责原制品关联和门禁。
- 类型、必填、选项校验在 Server 执行，不能依靠直接调用 API 绕过。配置加载失败时页面保留错误，不退回无校验提交。
- 自由变量仍保留；与控件同名时以控件值为准。`event`、`fork` 不可作为参数名；`deploy_to` / `deploy_task` 只允许用于 deployment。
- 不增加生产审批、风险确认、全局幂等或恢复机制；原 `Allow deployments`、权限、原门禁和具名确认继续生效。

## API 与安装边界

仓库参数 CRUD 使用 `/api/repos/{repo_id}/parameters`，运行仍使用原来的 manual 与 Deploy API。参数不是 Secrets，不能用于存放口令或令牌。

Server 新增 `parameters` 表，唯一键为仓库、运行入口、参数名。沿用官方 migration 与 SQLite 持久化；官方 Agent 不需要增加表单解释逻辑。验证和安装必须使用独立数据库与端口，不能把 fork 直接指向现有 V3 或官方测试实例的数据目录。

当前能力不等于 Go `next` 接入已经完成。SDK / connector 继续走本地通道。
