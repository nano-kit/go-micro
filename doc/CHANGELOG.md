# Changelog

## Unreleased

### Added

* `api/handler/rpc/stream.go` 修复 micro api 对 stream 流式调用的支持
* `api/handler/rpc/rpc.go` 处理 X-Forwarded-For 头，并根据 Client IP 作一致性hash选下游节点
* `api/handler/rpc/rpc_test.go` 设计测试用例，看一致性hash的效果
* `api/server/http/http/go` 去掉 CombinedLoggingHandler

修改记录在 https://github.com/aclisp/sims/commits/master/pkg/go-micro

## v2.9.1 - 2020-07-03
