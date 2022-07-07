# HPC调度系统中间件
# 简介
该项目主要用于将slurm等调度系统的队列、作业等相关信息，采用非侵入试的方式转换为方便传输和解析的json数据。  
采用此接口中间件可以在不改变原有调度系统的前提下，方便的操作调度系统，并以结构化数据的形式展现给客户端程序，大大简化了开发者操作操作调度和解析想用数据的流程。   
使用该中间件程序可以快速实现一个可视化的集群作业管理平台。  

## 该项目的优点
1、根据配置文件定时增量更新作业信息  
2、可根据配置文件自定义任何操作指令  
3、采用ssh连接的形式与调度系统进行交互，完全独立于集群系统  
4、不同指令的响应结果可根据配置文件灵活修改  
5、理论上秩序修改配置文件即可适配多种作业调度系统


## 下一步计划
1、根据已有的功能设计可视化界面  
2、开发多个调度系统间的协调调度功能  
3、集成文件相关的功能  

## 初始化连接接口 
接口地址： /schedule/v1/init
请求类型：POST
请求参数类型： json
请求参数：
```json
{
  "username": "test", 
  "password": "*****",
  "schedule_id": 1,
  "schedule_name": "slurm",
  "path": "./",
  "session_id": "****"
}
```
username: Linux用户名  
password：密码  
schedule_id：调度系统id，要与配置文件中填写的一致  
schedule_name: 调度系统名称，与配置文件中的调度系统名称一致（该参数与schedule_id两个参数二选一即可，schedule_id优先级大于schedule_name）  
path：默认计入的工作目录，可为空
session_id: 该连接的唯一性ID，可有客户端自由控制

正确响应数据：   
```json
{
    "code": 200,
    "data": {
      "ScheduleID": 1,
      "ConnID": "*****************"
    },
    "msg": "初始化客户端连接成功！"
}
```
错误响应数据：
```json
{
  "code": 506,
  "data": "",
  "msg": "初始化客户端连接异常！"
}
```


## 执行命令接口
接口地址：/schedule/v1/exec  
请求类型：POST  
请求参数类型：json

请求参数：
```json
{
  "conn_id": "123456789",
  "cmd": "sinfo",
  "schedule_id": 1,
  "command_id": "3",
  "args": ""
}
```
conn_id: 连接ID，为init接口相应的数据
cmd: 命令，需要与配置文件中保持一致
schedule_id: 调度系统id需要与配置文件一致
command_id: 命令id，与配置文件保持一致
args：命令参数

响应参数：
```json
{
  "code": 200,
  "data": "",
  "msg": "操作执行成功！"
}
```
