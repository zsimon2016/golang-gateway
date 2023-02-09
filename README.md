# golang-gateway
### 分布式微服务网关
这是一个分布式微服务网关，由RPC作为核心构建的微服务访问入口、治理中心。

本网关选择以etcd而没有以nacos作为注册中心是因为与go更加契合，同时占用的内存更低，在使用之前请自行安装好etcd，默认使用本地的2379作为注册中心端口，如果自定义注册中心，则只需要加上启动参数。-EtcdAddr ip:port。ip 和port自行替换。

### 监听  
默认监听9092端口，作为http服务的端口，如果需要使用https请与nginx配合使用。

### 启动  
```go run service.go ```  
### 参数说明:  
- basePath  etcd 服务注册路径，一般无需修改，各个微服务也是注册在这个路径下的。

- configPath  配置文件路径，默认./config.json 可以自行指定。这个文件内容其实是可以用其他配置中心所替代，后期会有升级。原来只需要一个nacos就可以，因为轻量化，把配置中心去除了。配置文件中的gatewayAddress、services_list、redis字段名请不要修改。services_list中的key是微服务向注册中心注册的服务名称，注意val的首字母应该大写。
