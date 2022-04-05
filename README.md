# 一个简单的NAT穿透用udp反向代理工具

## 使用场景
- 两个在NAT后面的设备会尝试穿透NAT并进行p2p连接。
- 分为服务器与客户端。客户端负责发起连接到服务器，并在本机开一个udp监听端口监听其它流量并将他们转发到服务端。服务端会接受客户端连接，并将客户端提交的流量转发至目标服务器。
- 全锥NAT可以和任何对端连接；端口限制型可以与端口限制型和全锥型对端连接；对称型只能与全锥型对端连接。

## 特性
- 服务器和客户端会创建一个stun连接用来确定自己的公网ip和端口
- 服务端可以主动发起连接来连接客户端，这在端口限制型服务端和客户端、全锥型客户端连接非全锥型客户端时是必要的
- 只能转发udp数据，如果要转发tcp数据可以配合kcp或者OpenVPN使用。
- 为了性能考虑，可以选择不会载荷进行加密和校验，也可以选择不会载荷进行校验。这在配合自带加密认证校验的协议的时候可以考虑，例如kcp和OpenVPN

## 安全性
- 零信任：除非客户端提交了正确的数据，否则服务端不会向客户端发送任何数据（主动连接除外）
- 抗重放攻击：为每个udp报文进行标号，接收到序列号小的报文则丢弃
- 连接创建报文会带有当前时间戳（时区无关），用于抗连接创建报文的重放攻击。连接成功后会使用新的密钥。
- 加密算法选NONE则不加密，许多安全特性会失效。


## 使用方法
下载可执行文件，或者自己编译一个。
命令行参数：
- -c 配置文件路径，默认config.json
- -l http api/webui 监听地址，默认127.0.0.1:3480
- -r Golang运行时调试工具的http api监听地址

配置文件路径指向的文件可以不存在，程序也能正常启动。
进入webui之后可以开始配置，配置完成后点击保存配置即可生成配置文件。
程序再次启动时会自动加载指定的配置文件。

## 替换前端静态资源
下载go-bindata-assetfs
```
go install github.com/go-bindata/go-bindata/...
go install github.com/elazarl/go-bindata-assetfs/...
```
下载源码，替换static文件夹里的前端静态文件，运行命令
```
go generate
go build -o uurp.exe  # Windows
go build -o uurp      # 其它平台
```

使用代码
- https://github.com/ccding/go-stun

萌新脸滚键盘作品

感觉写的挺不规范的，凑合用吧。有安全性或是其它bug欢迎提出来
