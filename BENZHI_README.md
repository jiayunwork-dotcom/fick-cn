# fick-cn：Go 一维 Fick 扩散 Crank–Nicolson Web 服务（剖面步进 + 通量核算 + 前端控制台）

给定杆长、扩散系数 D、网格与边界，用 Crank–Nicolson 推进浓度场，提供 `/api/step`、`/api/flux` 与嵌入网页。

## 构建 / 运行 / 测试

```text
go build ./...
./fick-cn -http :8080
curl -s http://127.0.0.1:8080/api/example
go run . step example/closed-rod.json
go test ./...
```

## 评测镜像

- `benzhi.Dockerfile`
- `build_benzhi_docker.sh`
- `BENZHI_README.md`（本文件）

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh <image-name> linux/amd64
docker run -d -P --name fick-cn-b14 <image-name>:latest
curl -s http://127.0.0.1:$(docker port fick-cn-b14 8080 | cut -d: -f2)/api/example
docker rm -f fick-cn-b14
```
