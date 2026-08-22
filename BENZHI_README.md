# fick-cn — 一维 Fick 扩散 Crank–Nicolson 核算

fick-cn 是命令行一维 Fick 扩散核算工具：输入杆长、扩散系数 D、网格与边界条件，用 Crank–Nicolson 三对角推进浓度场，打印剖面、总质量与近稳态判定。

## 构建 / 运行 / 测试

```text
go build ./...
go run . step example/closed-rod.json   # CLI：核算闭杆算例并打印剖面、总质量、近稳态
go test ./...
```

## 评测镜像

本目录评测专用文件（勿覆盖项目自带 Dockerfile/README）：

- `benzhi.Dockerfile`
- `build_benzhi_docker.sh`
- `BENZHI_README.md`（本文件）

两种架构都要构建并进容器验证：

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh <image-name> linux/arm64
./build_benzhi_docker.sh <image-name> linux/amd64
docker run -it <image-name>:latest
```

进容器后运行 `go build ./... && go test ./...`，再用 `go run . step example/closed-rod.json` 验证 CLI 输出总质量守恒与剖面峰值下降。
