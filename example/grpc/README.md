```
grpcurl -plaintext -d @ localhost:8081 hello.Hello.Echo <<EOM
{
    "msg": "hello world"
}
EOM
```
