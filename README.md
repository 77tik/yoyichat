# YoyiChat!!! 

### yoyichat 是一个使用go实现的轻量级im系统

### 架构设计
#### 登陆功能：
+ api层：基于gin提供http接口服务，主要任务是调用logic层注册在etcd的登陆方法
+ logic层：将登陆方法注册进rpcx框架内，主要是在redis中保存会话与用户元信息

#### 单聊信息发送：
+ api层：提供单聊信息发送接口给用户，自己则是调用logic层注册的单聊方法
+ logic层：从请求中拿到目标用户的id，拿到我们在redis保存了以用户id为前缀对应的serverId，构造请求，放入task层的消息队列中
+ task层：拿到了上面logic层放入消息队列中的请求，task层会启动消费者来消费队列中的请求
  + task层负责一个全局的用于调用connect层rpc方法的客户端rclient
  + 这个rclient内部有一个map，根据serverId关联 专门调用connect层rpc方法的xclient
  + 我们解析logic层传来的请求中得到serverId，再去map查找，发现了我们的目标xclient（可能有多个，我们每次都+1，不会一直都是一个xclient处理）
  + 然后让这个xclient帮我们调用connect层的rpc单聊方法
  + 多个connect层实例，方便水平扩展，一个serverId可以对应多个，idx保证负载均衡
  + 疑问2:etcd一个key能注册多个service？