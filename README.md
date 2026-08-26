# fick-cn — 一维 Fick 扩散 Crank–Nicolson 核算

fick-cn 是一维 Fick 扩散 Crank–Nicolson 核算服务：你给出杆长、扩散系数 D、网格、边界（Dirichlet 钉值或 Neumann 无通量）与初值，它用 θ=1/2 隐式平均在均匀网格上推进浓度场，报告剖面、梯形总质量、近稳态，以及面通量 J=-D ∂c/∂x。可用 HTTP（`-http :8080`，`/api/step` `/api/flux`）或 CLI 子命令。能力边界：只做均匀网格上一维线性扩散，不含对流、多组分或反应项。

- 输入：JSON 算例文件（`length`/`diffusivity`/`nodes`/`dt`/`t_end`/`boundary_left`/`boundary_right`/`initial`），示例见 `example/`。
- 输出：若干时刻的浓度剖面、总质量、峰值、近稳态判断；`steady` 子命令另给出解析稳态剖面。
- 非法输入（D≤0、L≤0、网格数 <3、Δt≤0、未知 JSON 字段、未知边界类型）一律报 stderr 并非零退出。

## 用法

在仓库根目录运行：

```text
go run . -http :8080
curl -s http://127.0.0.1:8080/api/example
go run . step example/closed-rod.json
```

`closed-rod.json` 是一个两端无通量、初值中央有一块脉冲的闭杆：总质量被钉死在初始值（相对漂移在 1e-12 量级），峰值随时间下降，最终趋向初值平均。

其余子命令：

```text
go run . steady example/dirichlet-rod.json   # 打印末态剖面与解析稳态剖面对比
go run . check  example/closed-rod.json      # 内置交叉核验：质量守恒、D×4 与 t/4 缩放、放大因子、稳态
go run . help                                # 用法说明
```

算例文件字段：

| 字段 | 含义 | 必填 |
|------|------|------|
| `length` | 杆长 L | 是（>0） |
| `diffusivity` | 扩散系数 D | 是（>0） |
| `nodes` | 网格节点数 N | 是（≥3） |
| `dt` | 时间步长 | 是（>0） |
| `t_end` | 目标时间 | 是（>0） |
| `boundary_left.kind` | `dirichlet`（附 `value`）或 `neumann` | 是 |
| `boundary_right.kind` | 同上 | 是 |
| `initial.kind` | `pulse`/`uniform`/`gaussian`/`linear`/`zero` | 是 |

## 关键约定

- **离散格式**：内部点满足 (c^{n+1}−c^n)/Δt = (D/2)(c_xx^{n+1}+c_xx^n)，两端无通量用镜像点 c_{-1}=c_1 处理，Dirichlet 端直接钉值。
- **总质量**：梯形积分 M = h(c_0/2 + c_1 + … + c_{N-2} + c_{N-1}/2)。无通量两端时该离散质量在 θ 平均格式下精确守恒。
- **稳态**：两端都钉死不同值时长时间极限为连接两端值的直线；两端无通量时极限为初值平均；单端钉死时另一端无通量最终被水库灌满为钉死值。
- **稳定性**：θ=1/2 的放大因子对任意 Fourier 模态与任意步长满足 |g|≤1；不显式前向欧拉冒充无条件稳定。
- **D–t 缩放**：c(x,t;D)=c(x,t/4;4D)，剖面只依赖乘积 D·t。
- **边界质量账**：Dirichlet 端有通量、Neumann 端通量为零，每步质量增量必须等于两端通量之和。

## 构建与测试

```text
go build ./...
go test ./...
```

HTTP：`POST /api/step` 提交与 CLI 相同的问题 JSON，返回质量、峰值、两端通量与 Fourier 数；`POST /api/flux` 对给定剖面求面通量。

## 许可

MIT，见 [LICENSE](./LICENSE)。
