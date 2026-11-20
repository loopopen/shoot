# v1.1

## [ ] Mapper 映射方法支持泛型

针对 Go 1.27 的泛型方法语法，让 Mapper 上的映射方法可以声明自己的类型参数。

- 方法名和参数列表之间可以有类型参数，例如 `func (m Mapper) Map[T any](v T) string`。
- 生成映射时能按字段类型实例化这些方法，类型参数可推断时不必手写类型实参。
- 现有单参数、单返回值、非泛型映射方法保持可用。

## [ ] enum 支持 case

`enum` 生成字符串时支持大小写转换。

- 增加 case 选项，至少覆盖 upper case，并一并考虑 lower、snake、camel 这类常见形式。
- `String`、JSON、Text、SQL 使用同一套转换结果。
- 未指定 case 时，字符串仍是去掉类型前缀后的常量名。

## [ ] restclient 抽象

把执行 HTTP 的部分抽成后端，生成代码可以在标准库 `http.Client` 和 resty client 之间切换。

先判断能否在 v1 保持兼容：

- 兼容意味着用户侧不用改：`RestClient`、`NewRest`、`ConfigHTTPClient`、`shoot.Response`、中间件和已生成的方法签名都保持不变，resty 只作为内部后端。
- 现有实现绑在 `*http.Client` 上：`RestClient.ConfigHTTPClient`、生成结构体里的 `client` 字段、`retry.Do`，以及 `RestConf.BuildMiddleware` 返回的 `http.RoundTripper`。resty 能包一层 `*http.Client` 并复用同一个 `RoundTripper`，这条路径有机会兼容。
- 如果要换成 resty 自己的重试、中间件或响应类型，公开接口对不上，就不要放进 v1.1，改到 v2 做。

## [x] map 复制目标字段注释

`map` 时可以把目标字段的注释复制到对应的源字段，用一个选项控制，取值为 `on`、`off`、`force`。

- `on`：复制。源字段已经有注释时保留原注释。默认值。
- `off`：不复制。
- `force`：强制覆盖，用目标字段的注释替换源字段上已有的注释。目标字段存在但没有注释时，清除匹配源字段的注释。
