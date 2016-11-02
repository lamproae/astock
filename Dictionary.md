# CORS : Cross-Origin Resource Sharing
    跨站资源共享标准通过新增一系列的HTTP头，让服务器能声明哪些来源可以通过浏览器访问该服务器上的资源。 另外， 对那么些会对服务器造成破坏性影响的HTTP请求方法（特别是GET以外的HTTP方法， 或者搭配某些MIMIE类型的POST请求）， 标准强烈要求浏览器必须先以OPTIONS请求方式发送一个预请求（preflight request）， 从而获知服务器端对跨源请求所支持的HTTP方法。在确认服务器允许该跨源请求的情况下，以实际的HTTP请求方法发送那个真正的请求。服务器端也可以通知客户端，是不是需要随同请求一起发送信用信息（包括Cookies和HTTP认证相关数据）。
