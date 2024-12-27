# redis-cli-hook
- hook req-ack，允许预处理数据，比如自定义编解码codec
- 只支持一问一答(不支持 monitor等流式协议)

# Usage
```
./redis-cli-hook -p 30302 -a lovengame
```

# Output
```
127.0.0.1:40637> hkeys id
req {"Type":"*","Value":[{"Type":"$","Value":"hkeys"},{"Type":"$","Value":"id"}]}
ack {"Type":"*","Value":[{"Type":"$","Value":"logic"},{"Type":"$","Value":"log"},{"Type":"$","Value":"resourcelog"}]}
1) "logic"
2) "log"
3) "resourcelog"
127.0.0.1:40637> hgetall name
req {"Type":"*","Value":[{"Type":"$","Value":"hgetall"},{"Type":"$","Value":"name"}]}
ack {"Type":"*","Value":[{"Type":"$","Value":"刘宜"},{"Type":"$","Value":"1125899906842645"},{"Type":"$","Value":"嵇望娅"},{"Type":"$","Value":"1125899906842633"},{"Type":"$","Value":"昝甜嫱"},{"Type":"$","Value":"1125899906842624"},{"Type":"$","Value":"褚玉岚"},{"Type":"$","Value":"1125899906842631"},{"Type":"$","Value":"钦岚昀"},{"Type":"$","Value":"1125899906842663"}]}
 1) "\xe5\x88\x98\xe5\xae\x9c"
 2) "1125899906842645"
 3) "\xe5\xb5\x87\xe6\x9c\x9b\xe5\xa8\x85"
 4) "1125899906842633"
 5) "\xe6\x98\x9d\xe7\x94\x9c\xe5\xab\xb1"
 6) "1125899906842624"
 7) "\xe8\xa4\x9a\xe7\x8e\x89\xe5\xb2\x9a"
 8) "1125899906842631"
 9) "\xe9\x92\xa6\xe5\xb2\x9a\xe6\x98\x80"
10) "1125899906842663"
```
