# redis-cli-hook
- hook req-ack，允许预处理数据，比如自定义编解码codec
- 只支持一问一答(不支持 monitor等流式协议)

# Usage
```
./redis-cli-hook -p 30302 -a xx
```

# Output
```
127.0.0.1:38283> hkeys id
[Redis-CLI] 2025/01/06 17:22:17 执行命令: hkeys, 参数: [id]
[Redis-CLI] 2025/01/06 17:22:17 命令结果: {"type":"*","value":[{"type":"$","value":"logic"},{"type":"$","value":"log"},{"type":"$","value":"resourcelog"}]}
1) "logic"
2) "log"
3) "resourcelog"
127.0.0.1:38283> hkeys globaldata
[Redis-CLI] 2025/01/06 17:22:22 执行命令: hkeys, 参数: [globaldata]
[Redis-CLI] 2025/01/06 17:22:22 命令结果: {"type":"*","value":[{"type":"$","value":"save_time"},{"type":"$","value":"server_id"},{"type":"$","value":"boss"},{"type":"$","value":"coliseum"}]}
1) "save_time"
2) "server_id"
3) "boss"
4) "coliseum"
```