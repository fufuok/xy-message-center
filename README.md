# XY.MessageCenter.DingTalk (迅游消息中心)

## 功能

- [x] 钉钉
  - [x] `/v1/dd/chat/create` 创建会话, 即建内部群, 参考:
    - https://developers.dingtalk.com/document/app/create-group-session
    - 参数: `name` `owner` `useridlist`
    - 主要为了得到 `chatid` 下面的接口要用.
  - [x] `/v1/dd/chat/send` 发送群消息, 参考:
    - https://developers.dingtalk.com/document/app/send-group-messages
    - 只关注参数: `chatid` `msg`
    - `msg` 支持的消息格式和参数见:
      - https://developers.dingtalk.com/document/app/message-types-and-data-format
  - [x] `/v1/dd/topapi/message` 工作通知消息, 参考:
    - https://developers.dingtalk.com/document/app/asynchronous-sending-of-enterprise-session-messages
    - 只关注参数: `userid_list` `dept_id_list` `to_all_user` `msg`
    - 注意: 前 3 个参数必须有一个不为空.
    - `msg` 与上面的接口相同.
    - 注: 图片, 语音, 文件需要先上传到钉钉, 使用 `media_id` 发出消息
  - [x] `/v1/dd/media/upload` 媒体文件上传, 参考:
    - https://developers.dingtalk.com/document/app/upload-media-files
    - 参数: `type`, `media`
    - 返回值 `data['data']` 即为 `media_id`

## 依赖

见: go.mod

## 结构

    .
    ├── common      公共结构定义和方法, 全局变量
    ├── conf        配置文件目录
    ├── controller  控制器, 路由
    ├── doc         开发文档
    ├── log         日志目录
    ├── master      服务端程序初始化
    ├── middleware  Web 中间件
    ├── service     应用逻辑
    ├── model       模型, 数据交互
    ├── util        工具类库
    └── main.go     入口文件

## 说明

1. 环境变量及加密小工具见: `tools`
2. 运行 `./xymessagecenter` 默认使用配置文件目录下 `../etc/xymessagecenter.json`
3. 可以指定配置文件运行 `./xymessagecenter -c /mydir/conf.json`
4. 自动后台运行并守护自身, `Warn` 和守护日志在 `log/daemon.log`, 错误日志按天存放于 `log` 并发到 ES
5. Redis 和系统状态访问: http://api.domain:27779/sys/status JSON 格式, 可用于报警
6. 心跳检查地址: http://api.domain:27779/heartbeat 返回字符串 `OK`, `/ping` 返回字符串 `PONG`

## 接口示例

### 1. 接口说明

接口完全按钉钉开发平台文档设计, 参数名大致相同, 详见接口文档.

### 1. 发送工作消息

```http
POST /v1/dd/topapi/message HTTP/1.1
Host: api.domain:27779
Content-Type: application/json

{
    "userid_list": "0632500561850620,0632500561850621",
    "msg": {
        "msgtype": "text",
        "text": {
            "content": "工作消息测试"
        }
    }
}
```

### 2. 应答

```json
{
    "id": 1,
    "ok": 1,
    "code": 0,
    "msg": "",
    "data": null,
    "count": 0
}
```

### 3. 错误应答示例

```json
{
    "id": 0,
    "ok": 0,
    "code": 1,
    "msg": "群消息发送失败",
    "data": "[34014] 会话消息的json结构无效或不完整",
    "count": 0
}
```

`ok` 或 `id` 为 1 表示成功, 0 表示失败, `code` 相反, 0 表示成功, 1 表示失败







*ff*