# fick-cn — 一维 Fick 扩散 Crank–Nicolson 核算

fick-cn 是一个命令行一维 Fick 扩散核算工具。你给出一根杆的长度、扩散系数 D、网格规模、边界条件（两端可分别钉死为 Dirichlet 或设为无通量 Neumann）与初始浓度场，它用 Crank–Nicolson 隐式平均（θ=1/2）在均匀网格上按时间步进浓度场，并报告各时刻剖面、梯形总质量与是否已接近稳态。能力边界：只做均匀网格上的一维线性扩散，不涉及对流项、多组分相互作用或反应项；显式前向欧拉不被当作无条件稳定方案使用。

- 输入：JSON 算例文件（`length`/`diffusivity`/`nodes`/`dt`/`t_end`/`boundary_left`/`boundary_right`/`initial`），示例见 `example/`。
- 输出：若干时刻的浓度剖面、总质量、峰值、近稳态判断；`steady` 子命令另给出解析稳态剖面。
- 非法输入（D≤0、L≤0、网格数 <3、Δt≤0、未知 JSON 字段、未知边界类型）一律报 stderr 并非零退出。

## 用法

在仓库根目录运行：

```text
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
go build ./...       # 编译（纯标准库，无第三方依赖）
go test ./...        # 全部单元测试（spec / mesh / boundary / operator / field / advance / check）
```

## 许可

MIT，见 [LICENSE](./LICENSE)。
