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
  + logic 层处理加入房间时，会将 yoyichat_2918 : s01 即用户key ：connect层id作关联，加入redis中，类型为string
    + 并不只有用户，在关于群聊相关时，都会先从redis中获取 yoyichat_room_r01 即房间号，成员id作为房间号这个hash的字段，加入房间是就是加入字段
    + 当然还有一个在线人数，redis中存的是string，yoyichat_room_online_count_r01 可以获取roomid为 r01 的在线人数
  + connect层 监听连接建立，这是让客户端直接获取信息的最近的途径，有客户端连过来的时候就会调用logic层的Connect方法，断联则会调用DisConnect方法
  + 现在的当务之急是这个api层和connect层似乎是平行的，我可以对api层发号施令，但是connect层这条是怎么来的
  + 当然是因为rpcx建立连接是建立在调用方法上的啊，只要调用方法就会建立连接，这也就意味着我们做client的时候需要调用api层和connect层的方法，才能建立连接