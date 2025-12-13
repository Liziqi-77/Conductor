# vLLM 引擎层架构分析

## 1. 引擎层整体架构

引擎层是vLLM的核心执行层，负责管理请求的生命周期、调度执行、处理输入输出。主要包含以下组件：

1. **AsyncLLM** - 异步LLM引擎客户端（主要接口）
2. **LLMEngine** - 同步LLM引擎（向后兼容）
3. **EngineCore** - 引擎核心，包含调度和执行循环
4. **EngineCoreProc** - 进程包装的EngineCore，支持多进程执行
5. **EngineCoreClient** - 引擎核心客户端，处理进程间通信
6. **InputProcessor** - 输入处理器
7. **OutputProcessor** - 输出处理器

## 2. 核心组件详解

### 2.1 AsyncLLM - 异步LLM引擎客户端

**文件位置**：`vllm/v1/engine/async_llm.py`

**类定义**：`AsyncLLM(EngineClient)`

**功能**：
- 提供异步接口给服务层
- 管理请求的输入处理和输出收集
- 与EngineCore进程通信
- 处理请求的生命周期

#### 2.1.1 初始化方法

**`__init__()`**
```python
def __init__(
    self,
    vllm_config: VllmConfig,
    executor_class: type[Executor],
    log_stats: bool,
    usage_context: UsageContext = UsageContext.ENGINE_CONTEXT,
    ...
)
```

**功能**：
- 初始化配置和组件
- 创建InputProcessor、OutputProcessor
- 创建EngineCoreClient（多进程客户端）
- 初始化统计日志管理器
- 启动输出处理循环

**关键步骤**：
1. 初始化tokenizer（如果未跳过）
2. 创建InputProcessor处理输入
3. 创建OutputProcessor处理输出
4. 创建EngineCoreClient（异步多进程客户端）
5. 创建StatLoggerManager（如果启用统计）
6. 启动输出处理任务

#### 2.1.2 添加请求方法

**`add_request()`**
```python
async def add_request(
    self,
    request_id: str,
    prompt: EngineCoreRequest | PromptType,
    params: SamplingParams | PoolingParams,
    arrival_time: float | None = None,
    lora_request: LoRARequest | None = None,
    ...
) -> RequestOutputCollector
```

**功能**：
- 添加新请求到引擎
- 处理输入转换
- 支持n>1的并行采样（fan-out子请求）
- 返回输出收集器

**执行流程**：
1. 检查引擎是否出错
2. 创建RequestOutputCollector用于收集输出
3. 如果prompt不是EngineCoreRequest，通过InputProcessor转换
4. 如果n>1，创建多个子请求（fan-out）
5. 调用`_add_request()`添加到引擎

**`_add_request()`**
```python
async def _add_request(
    self,
    request: EngineCoreRequest,
    prompt: str | None,
    parent_req: ParentRequest | None,
    index: int,
    queue: RequestOutputCollector,
)
```

**功能**：
- 将请求添加到OutputProcessor（当前进程）
- 将请求添加到EngineCore（后台进程）
- 记录请求日志

#### 2.1.3 生成方法

**`generate()`**
```python
async def generate(
    self,
    prompt: EngineCoreRequest | PromptType,
    sampling_params: SamplingParams,
    request_id: str,
    ...
) -> AsyncGenerator[RequestOutput, None]
```

**功能**：
- 主要的生成接口，被API服务器调用
- 创建异步生成器返回RequestOutput
- 后台任务处理输出

**执行流程**：
1. 验证参数（如prompt_logprobs与kv_sharing_fast_prefill的兼容性）
2. 调用`add_request()`添加请求
3. 返回异步生成器，从RequestOutputCollector获取输出
4. 后台output_handler循环从EngineCore获取输出并放入collector

### 2.2 EngineCore - 引擎核心

**文件位置**：`vllm/v1/engine/core.py`

**类定义**：`EngineCore`

**功能**：
- 管理调度器（Scheduler）
- 管理模型执行器（Executor）
- 执行推理循环
- 处理KV缓存
- 协调请求的执行

#### 2.2.1 初始化方法

**`__init__()`**
```python
def __init__(
    self,
    vllm_config: VllmConfig,
    executor_class: type[Executor],
    log_stats: bool,
    executor_fail_callback: Callable | None = None,
)
```

**功能**：
- 初始化所有核心组件
- 设置KV缓存
- 创建调度器
- 创建模型执行器

**关键步骤**：
1. 加载插件
2. 创建模型执行器（Executor）
3. 初始化KV缓存（`_initialize_kv_caches()`）
4. 创建结构化输出管理器
5. 创建调度器（Scheduler）
6. 初始化多模态注册表和缓存
7. 设置批处理队列（如果支持流水线并行）

#### 2.2.2 KV缓存初始化

**`_initialize_kv_caches()`**
```python
def _initialize_kv_caches(
    self, vllm_config: VllmConfig
) -> tuple[int, int, KVCacheConfig]
```

**功能**：
- 分析模型需要的KV缓存规格
- 分析可用GPU内存
- 计算可分配的GPU块数量
- 初始化KV缓存

**执行流程**：
1. 从模型执行器获取KV缓存规格
2. 分析可用GPU内存（通过内存分析或同步）
3. 计算KV缓存配置
4. 初始化KV缓存
5. 返回块数量和配置

#### 2.2.3 添加请求方法

**`add_request()`**
```python
def add_request(self, request: Request, request_wave: int = 0):
```

**功能**：
- 添加请求到调度器
- 验证请求参数
- 处理数据并行场景的wave

**执行流程**：
1. 验证request_id类型
2. 如果是pooling请求，验证任务类型
3. 检查KV传输参数
4. 调用调度器的`add_request()`

#### 2.2.4 核心执行循环

**`step()`**
```python
def step(self) -> tuple[dict[int, EngineCoreOutputs], bool]:
```

**功能**：
- 执行一步推理
- 调度请求、执行模型、更新状态

**执行流程**：
1. 检查是否有待处理的请求
2. 调用调度器的`schedule()`获取调度输出
3. 调用执行器的`execute_model()`执行模型（非阻塞）
4. 获取语法掩码（如果启用）
5. 等待模型输出完成
6. 如果模型输出为None，调用`sample_tokens()`
7. 调用调度器的`update_from_output()`更新状态
8. 返回引擎核心输出和执行标志

**`step_with_batch_queue()`**
```python
def step_with_batch_queue(
    self,
) -> tuple[dict[int, EngineCoreOutputs] | None, bool]:
```

**功能**：
- 使用批处理队列执行（支持流水线并行）
- 优先填充批处理队列
- 异步执行多个批次

**执行流程**：
1. 如果批处理队列未满且有请求，调度新批次
2. 将批次添加到队列（非阻塞）
3. 如果队列未满且最后一个批次未完成，返回None（继续调度）
4. 否则，等待队列中的批次完成
5. 更新调度器状态
6. 处理延迟的采样（如果有）

### 2.3 EngineCoreProc - 进程包装的EngineCore

**文件位置**：`vllm/v1/engine/core.py`

**类定义**：`EngineCoreProc(EngineCore)`

**功能**：
- 在后台进程中运行EngineCore
- 使用ZMQ进行进程间通信
- 处理输入队列和输出队列

#### 2.3.1 核心循环

**`run_busy_loop()`**
```python
def run_busy_loop(self):
```

**功能**：
- 引擎核心的主循环
- 处理输入队列和执行步骤

**执行流程**：
```python
while True:
    # 1) 处理输入队列直到有工作要做
    self._process_input_queue()
    
    # 2) 执行引擎步骤并返回输出
    self._process_engine_step()
```

**`_process_input_queue()`**
```python
def _process_input_queue(self):
```

**功能**：
- 处理来自客户端的请求
- 等待直到有工作要做

**执行流程**：
1. 如果引擎未运行且调度器无请求且批处理队列为空，等待输入
2. 从输入队列获取请求
3. 调用`_handle_client_request()`处理请求
4. 处理队列中剩余的所有请求

**`_process_engine_step()`**
```python
def _process_engine_step(self) -> bool:
```

**功能**：
- 执行引擎步骤
- 将输出放入输出队列

**执行流程**：
1. 调用`step_fn()`（step或step_with_batch_queue）
2. 将输出放入输出队列
3. 调用`post_step()`后处理
4. 返回是否执行了模型

**`_handle_client_request()`**
```python
def _handle_client_request(
    self, request_type: EngineCoreRequestType, request: Any
) -> None:
```

**功能**：
- 处理来自客户端的请求
- 支持ADD、ABORT、RECONFIGURE等请求类型

### 2.4 EngineCoreClient - 引擎核心客户端

**文件位置**：`vllm/v1/engine/core_client.py`

**功能**：
- 提供与EngineCore进程通信的接口
- 支持同步和异步通信
- 处理进程间消息传递

#### 2.4.1 异步多进程客户端

**`AsyncMPClient`**
- 异步接口
- 使用ZMQ进行进程间通信
- 管理输入输出队列

**`add_request_async()`**
```python
async def add_request_async(self, request: EngineCoreRequest) -> None:
```

**功能**：
- 异步添加请求到引擎核心
- 发送请求到后台进程
- 确保输出队列任务运行

### 2.5 InputProcessor - 输入处理器

**文件位置**：`vllm/v1/engine/input_processor.py`

**类定义**：`InputProcessor`

**功能**：
- 处理输入转换
- 验证采样参数
- 处理多模态输入
- Tokenization和预处理

#### 2.5.1 核心方法

**`process_inputs()`**
```python
def process_inputs(
    self,
    request_id: str,
    prompt: PromptType,
    params: SamplingParams | PoolingParams,
    arrival_time: float | None = None,
    lora_request: LoRARequest | None = None,
    ...
) -> EngineCoreRequest
```

**功能**：
- 将输入转换为EngineCoreRequest
- 验证参数
- 处理多模态数据
- 处理结构化输出

**执行流程**：
1. 验证参数（`_validate_params()`）
2. 验证多模态UUID（如果提供）
3. 预处理输入（`_preprocess_inputs()`）
4. 处理结构化输出（如果启用）
5. 创建EngineCoreRequest对象
6. 返回请求对象

**`_validate_params()`**
```python
def _validate_params(
    self,
    params: SamplingParams | PoolingParams,
):
```

**功能**：
- 验证采样参数
- 检查logprobs、logit_bias等
- 验证支持的参数组合

**`_preprocess_inputs()`**
```python
def _preprocess_inputs(
    self,
    prompt: PromptType,
    params: SamplingParams | PoolingParams,
    lora_request: LoRARequest | None,
    ...
) -> ProcessorInputs
```

**功能**：
- 预处理输入数据
- 处理多模态数据
- 处理编码器-解码器输入
- 返回处理后的输入

### 2.6 OutputProcessor - 输出处理器

**文件位置**：`vllm/v1/engine/output_processor.py`

**类定义**：`OutputProcessor`

**功能**：
- 将EngineCoreOutput转换为RequestOutput
- 管理请求状态
- 处理流式输出
- 处理logprobs和detokenization

#### 2.6.1 核心方法

**`add_request()`**
```python
def add_request(
    self,
    request: EngineCoreRequest,
    prompt: str | None,
    parent_req: ParentRequest | None,
    request_index: int,
    queue: RequestOutputCollector,
):
```

**功能**：
- 添加新请求到输出处理器
- 创建RequestState
- 初始化logprobs处理器和detokenizer

**`process_outputs()`**
```python
def process_outputs(
    self,
    outputs: dict[int, EngineCoreOutput],
    engine_core_timestamp: float,
    iteration_stats: IterationStats | None = None,
) -> OutputProcessorOutput
```

**功能**：
- 处理引擎核心输出
- 更新请求状态
- 生成RequestOutput
- 处理完成的请求

**执行流程**：
1. 遍历所有输出
2. 对每个输出，找到对应的RequestState
3. 更新请求状态（token、logprobs等）
4. 检查是否完成
5. 生成RequestOutput
6. 放入RequestOutputCollector
7. 返回需要中止的请求列表

## 3. 数据流和请求处理流程

### 3.1 请求添加流程

```
服务层调用AsyncLLM.add_request()
  ↓
InputProcessor.process_inputs() - 转换输入
  ↓
OutputProcessor.add_request() - 创建RequestState
  ↓
EngineCoreClient.add_request_async() - 发送到后台进程
  ↓
EngineCoreProc._handle_client_request() - 处理请求
  ↓
EngineCore.add_request() - 添加到调度器
  ↓
Scheduler.add_request() - 加入等待队列
```

### 3.2 执行循环流程

```
EngineCoreProc.run_busy_loop()
  ↓
_process_input_queue() - 处理输入
  ↓
_process_engine_step() - 执行步骤
  ↓
EngineCore.step() 或 step_with_batch_queue()
  ↓
Scheduler.schedule() - 调度请求
  ↓
Executor.execute_model() - 执行模型
  ↓
Executor.sample_tokens() - 采样token
  ↓
Scheduler.update_from_output() - 更新状态
  ↓
输出放入output_queue
  ↓
EngineCoreClient接收输出
  ↓
OutputProcessor.process_outputs() - 处理输出
  ↓
RequestOutputCollector.put() - 放入收集器
  ↓
服务层从生成器获取RequestOutput
```

### 3.3 输出处理流程

```
EngineCore输出 → EngineCoreOutput
  ↓
OutputProcessor.process_outputs()
  ↓
更新RequestState（token、logprobs等）
  ↓
检查完成条件
  ↓
生成RequestOutput
  ↓
RequestOutputCollector.put()
  ↓
AsyncLLM的output_handler循环
  ↓
从collector获取并yield给调用者
```

## 4. 关键设计模式

### 4.1 异步处理
- AsyncLLM使用asyncio提供异步接口
- 后台任务处理输出
- 非阻塞的进程间通信

### 4.2 进程隔离
- EngineCore在独立进程中运行
- 使用ZMQ进行进程间通信
- 隔离模型执行和API服务

### 4.3 批处理队列
- 支持流水线并行
- 异步执行多个批次
- 最大化GPU利用率

### 4.4 状态管理
- RequestState管理每个请求的状态
- OutputProcessor维护请求状态映射
- 支持流式输出和完整输出

## 5. 关键数据结构

### 5.1 EngineCoreRequest
```python
class EngineCoreRequest:
    request_id: str
    prompt_token_ids: list[int] | None
    mm_features: list[MultiModalFeatureSpec] | None
    sampling_params: SamplingParams | None
    pooling_params: PoolingParams | None
    arrival_time: float
    lora_request: LoRARequest | None
    priority: int
```

### 5.2 RequestState
```python
class RequestState:
    request_id: str
    prompt: str | None
    prompt_token_ids: list[int] | None
    logprobs_processor: LogprobsProcessor | None
    detokenizer: IncrementalDetokenizer | None
    queue: RequestOutputCollector | None
    is_prefilling: bool
    # ...
```

### 5.3 RequestOutputCollector
```python
class RequestOutputCollector:
    output: RequestOutput | PoolingRequestOutput | Exception | None
    ready: asyncio.Event
    aggregate: bool  # 是否聚合流式输出
```

## 6. 总结

引擎层是vLLM的核心执行层，主要特点：

1. **异步设计**：使用AsyncLLM提供异步接口，支持高并发
2. **进程隔离**：EngineCore在独立进程中运行，隔离模型执行
3. **灵活调度**：通过Scheduler实现连续批处理
4. **高效通信**：使用ZMQ进行进程间通信
5. **状态管理**：RequestState和OutputProcessor管理请求生命周期
6. **流式支持**：支持流式输出和完整输出两种模式

这种设计使得vLLM能够高效地处理大量并发请求，同时保持低延迟和高吞吐量。
