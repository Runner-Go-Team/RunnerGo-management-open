# RunnerGo-management-open

```text
management 服务为RunnerGo后端接口
```

# 修改配置文件， open.yaml
```yaml
base:
  is_debug: false                         #是否开启debug
  domain: "https://open.runnergo.cc/"     #项目地址
  max_concurrency: 1000000                #最大并发数

http:
  port: 1234                             #management项目端口号

mysql:
  username: "****"                       #mysql账号
  password: "****"           #mysql密码
  host: "************"      #mysql地址
  port: 3306                #端口号
  dbname: "****"            #数据库名称
  charset: "utf8mb4"        #字符集

mongodb:
  dsn: "mongodb://****:****@127.0.0.1:1000/****"  #mongodb数据库DNS地址
  database: "runnergo_open"     #数据库名称
  pool_size: 20                 #连接数

jwt:
  issuer: "****"            #jwt使用者
  secret: "kp#test"         #jwt加密秘钥

clients:
  runner:
    run_api: "https://****/runner/run_api"          #调试api接口
    run_scene: "https://****/runner/run_scene"      #调试场景接口
    stop_scene: "https://****/runner/stop_scene"    #停止场景调试接口
    run_plan: "https://****/runner/run_plan"        #运行计划接口
    stop_plan: "https://****/runner/stop"           #停止计划接口

#两种日志使用方式都可
log:
  InfoPath: "/data/logs/RunnerGo/RunnerGo_management-info.log"    #操作日志目录文件
  ErrPath: "/data/logs/RunnerGo/RunnerGo_management-err.log"      #操作日志目录文件

proof:
  info_log: "/data/logs/RunnerGo/RunnerGo_management-info.log"    #操作日志目录文件
  err_log: "/data/logs/RunnerGo/RunnerGo_management-err.log"      #错误日志目录文件


redis:
  address: "127.0.0.1:6379"     #redis地址
  password: "apipost"           #redis密码
  db: 1                         #redis使用库           

redisReport:                    #报告使用redis
  address: "127.0.0.1:6379"     #redis地址
  password: "apipost"           #redis密码
  db: 1                         #redis使用库 

smtp:
  host: "smtpdm.aliyun.com"     #邮箱服务地址
  port: 123                     #端口号
  email: "*********"            #邮箱
  password: "*******"           #邮箱密码

inviteData:
  AesSecretKey:  "******"  #邀请链接加密密钥

canUsePartitionTotalNum: 2   #初始化压力机可使用分区
```


# 启动
```text
 配置完成后，在根目录./main启动management服务
```


## 开源部署
1. 配置环境变量
## 配置说明
| key                                | 是否必填 | 默认值                                                |                            说明 |
|:-----------------------------------|-----|----------------------------------------------------|------------------------------:|
| 本机配置                               ||||
| RG_IS_DEBUG                        | 否   | false                                              |                     是否开启debug |
| RG_DOMAIN                          | 否   |                                                    |                RunnerGo项目入口地址 |
| RG_MANAGEMENT_HTTP_PORT            | 否   | 30000                                              |                   manage项目端口号 |
| Mysql数据库                           ||||
| RG_MYSQL_HOST                      | 否   | 127.0.0.0                                          |                    Mysql数据库地址 |
| RG_MYSQL_USERNAME                  | 否   | root                                               |                     Mysql用户名称 |
| RG_MYSQL_PASSWORD                  | 否   |                                                    |                       Mysql密码 |
| RG_MYSQL_DBNAME                    | 否   | runnergo                                           |                    Mysql数据库名称 |
| RG_MYSQL_CHARSET                   | 否   | utf8mb4                                            |                      Mysql字符集 |
| RG_MYSQL_PORT                      | 否   | 3306                                               |                      Mysql端口号 |
| JWT网络令牌                            ||||
| RG_JWT_ISSUER                      | 否   | RunnerGo                                           |                         JWT账号 |
| RG_JWT_SECRET                      | 否   | RunnerGo#docker                                    |                         JWT密钥 |
| mongo数据库                           ||||
| RG_MONGO_DSN                       | 否   | mongodb://runnergo:123456@127.0.0.0:27017/runnergo |                   mongo数据库dsn |
| RG_MONGO_DATABASE                  | 否   | runnergo                                           |                  mongo使用数据库名称 |
| RG_MONGO_PASSWORD                  | 否   |                                                    |                    mongo数据库密码 |
| RG_MONGODB_POOL_SIZE               | 否   | 20                                                 |                    mongo数据库密码 |
| engine服务接口                         |    |                                                    |                               |
| RG_CLIENTS_ENGINE_RUN_API          | 否   | https://127.0.0.0:30000/runner/run_api             |                 engine服务-调试接口 |
| RG_CLIENTS_ENGINE_RUN_SCENE        | 否   | https://127.0.0.0:30000/runner/run_scene           |                 engine服务-调试场景 |
| RG_CLIENTS_ENGINE_STOP_SCENE       | 否   | https://127.0.0.0:30000/runner/stop_scene          |               engine服务-停止调试场景 |
| RG_CLIENTS_ENGINE_RUN_PLAN         | 否   | https://127.0.0.0:30000/runner/run_plan            |                 engine服务-运行计划 |
| RG_CLIENTS_ENGINE_STOP_PLAN        | 否   | https://127.0.0.0:30000/runner/stop                |                 engine服务-停止计划 |
| proof日志目录                          ||||
| RG_PROOF_INFO_LOG                  | 否   | /data/logs/RunnerGo/RunnerGo_management-info.log   |                  proof-操作日志地址 |
| RG_PROOF_ERR_LOG                   | 否   | /data/logs/RunnerGo/RunnerGo_management-err.log    |                  proof-错误日志地址 |
| Redis                              ||||
| RG_REDIS_ADDRESS                   | 否   | 默认：127.0.0.0:6379                                  |                    redis服务端地址 |
| RG_REDIS_PASSWORD                  | 是   |                                                    |                    redis服务端密码 |
| RG_REDIS_DB                        | 否   | 默认：0                                               |                      redis数据库 |
| Redis-报告专属redis                    ||||
| RG_REDIS_REPORT_ADDRESS            | 否   | 默认：127.0.0.0:6379                                  |                    redis服务端地址 |
| RG_REDIS_REPORT_PASSWORD           | 是   |                                                    |                    redis服务端密码 |
| RG_REDIS_REPORT_DB                 | 否   | 默认：0                                               |                      redis数据库 |
| SMTP-邮件配置                          ||||
| RG_SMTP_HOST                       | 否   |                                                    |                        邮件服务地址 |
| RG_SMTP_PORT                       | 是   |                                                    |                       邮件服务端口号 |
| RG_SMTP_EMAIL                      | 否   |                                                    |                          邮箱名称 |
| RG_SMTP_PASSWORD                   | 否   |                                                    |                          邮箱名称 |
| 邀请链接验证密钥                           ||||
| RG_INVITE_DATA_AES_SECRET_KEY      | 否   | qazwsxedcrfvtgby                                   | 邀请链接验证密钥（key 长度必须 16/24/32长度） |
| 普通日志目录                             ||||
| RG_LOG_INFO_PATH                   | 否   | /data/logs/RunnerGo/RunnerGo_management-info.log   |                        操作日志地址 |
| RG_LOG_ERR_PATH                    | 否   | /data/logs/RunnerGo/RunnerGo_management-err.log    |                        错误日志地址 |
| 初始化压力机可使用分区                        ||||
| RG_CAN_USE_PARTITION_TOTAL_NUM     | 否   | 2                                                  |                   初始化压力机可使用分区 |
| 压力机相关配置                            ||||
| RG_ONE_MACHINE_CAN_CONCURRENCE_NUM | 否   | 5000                                               |                单台压力机能快速负载的并发数 |
| RG_MACHINE_ALIVE_TIME              | 否   | 10                                                 |              压力机上报心跳超时时间，单位：秒 |
| RG_INIT_PARTITION_TOTAL_NUM        | 否   | 2                                                  |              初始化可用kafka分区数量设置 |
| RG_CPU_TOP_LIMIT                   | 否   | 65                                                 |             可参与压测的压力机cpu使用率上限 |
| RG_MEMORY_TOP_LIMIT                | 否   | 65                                                 |          可参与压测的压力机memory使用率上限 |
| RG_DISK_TOP_LIMIT                  | 否   | 55                                                 |            可参与压测的压力机disk使用率上限 |
| 一些独立的配置项                           ||||
| RG_DEFAULT_TOKEN_EXPIRE_TIME                  | 否   | 24                                                 |       默认用户登录token的失效时间（单位：小时） |
| RG_KEEP_STRESS_DEBUG_LOG_TIME                  | 否   | 1                                                  |        保留性能测试的debug日志时间（单位：月） |