// Package cloudclient 封装 Laravel 云平台 REST（按 platform 选 base URL）、各资源 Client 与响应 JSON 工具。
// 仅服务于公网 cloud_public；内网 cloud_intranet 使用独立包 platform/cloud_intranet/client 与独立 cloudapi，避免两平台代码耦合。
package cloudclient
