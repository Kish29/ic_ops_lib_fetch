struct BaseReq {
    1: string Name  // 请求调用方名称
    2: optional map<string, string> ReqExtra // 请求的额外信息
}

struct BaseResp {
    1: string StatusMsg // 响应的短语
    2: i64 StatusCode   // 错误码
    3: optional map<string, string> RespExtra // 请求的额外信息
}