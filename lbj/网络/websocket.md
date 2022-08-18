# WebSocket

Websocket是服务器与客户端之间的双向通信形式，通常用于建立聊天室或多人视频游戏，因为这些应用程序需要服务器与客户端之间的持续通信。

## SSE(Server-Sent Events)

SSE通过http(s)链接创建一个单向流。服务器端可以源源不断的向客户端推送数据，客户端意外断开后还可以自动重连，甚至设置重连间隔。

在http的基础上，似乎头部加上"Content-Type":"text/event-stream"，传输数据符合如下格式即可:

```
data: {\n
data: "foo": "bar"\n
data: }\n\n
```

go相关实现例子:

```go
//服务器端必须设置 Content-Type 为 text/event-stream 表明这是个 event 流。
//no-cache 主要是避免客户端读缓存，keep-alive 保持长链接
func handleSSE(w http.ResponseWriter, r *http.Request) {

    appId := r.URL.Query()["appId"]
    page := r.URL.Query()["page"]
    pageSize := r.URL.Query()["pageSize"]

    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("Access-Control-Allow-Origin", "*")

    flusher, ok := w.(http.Flusher)
    if !ok {
        log.Panic("server not support")
    }
    for i := 0; i < 10; i++ {
        time.Sleep(5*time.Second)
        fmt.Fprintf(w, "data: %d%s%s%s\n\n", i, appId[0], page[0], pageSize[0])
        flusher.Flush()
    }
    fmt.Fprintf(w, "event: close\ndata: close\n\n") // 一定要带上data，否则无效
}

func main() {
    http.Handle("/event", http.HandlerFunc(handleSSE))
    http.ListenAndServe(":8000", nil)
}
```

与websocket对比：

* SSE提供单向通信，Websocket提供双向通信；
* SSE是通过HTTP协议实现的，Websocket是单独的协议；
* 实现上来说SSE比较容易，Websocket复杂一些；
* SSE有最大连接数限制；
* WebSocket可以传输二进制数据和文本数据，但SSE只有文本数据；

与长轮询相比:实时性更强，如实时计数统计之类的功能。长轮询用在，如等待计算结果这种长耗时，不需要实时检查的任务上。