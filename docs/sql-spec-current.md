# SQL 开发规范

## 1. 适用范围

- 面向 MySQL / TiDB SQL，支持建表、改表、索引、视图、`INSERT`、`UPDATE`、`DELETE`、`REPLACE`、`ANALYZE TABLE` 等语句。
- 单独的 `SELECT` 不属于可审核语句；未支持的语句也会返回告警。
- 审核结果中的 `WARN` 表示规则未通过，调用方应按失败处理，不能仅以接口调用成功作为放行依据。
- 表结构、对象存在性、索引长度及 DML 影响行数等检查依赖目标数据库的实时元数据和 `EXPLAIN`。

## 2. 开发必须遵守

| 类别 | 当前要求 |
| --- | --- |
| 命名 | 表名、列名最长 64 个字符，只使用字母、数字和下划线。当前未启用关键字检查，但仍禁止在业务 SQL 中使用保留字命名。 |
| 建表 | 必须显式指定 `ENGINE=InnoDB`；必须显式指定 `utf8`、`utf8mb4` 或 `latin1` 字符集。若指定排序规则，其名称必须与字符集匹配。 |
| 主键 | 每张表必须有且只能有一个主键，主键最多 2 列，主键列必须显式 `NOT NULL`。当前不强制主键为 `BIGINT UNSIGNED AUTO_INCREMENT`。 |
| 列 | 除 `BLOB`、`TEXT`、`JSON`、`GEOMETRY` 外，列必须设置默认值；上述类型不得设置默认值。`CHAR` 长度超过 64 会告警，建议改为 `VARCHAR`。 |
| 索引 | 二级索引（含唯一索引）最多 12 个，每个最多 8 列；禁止直接对 `BLOB` / `TEXT` 列建索引；重复索引名会告警。 |
| DML | `UPDATE`、`DELETE` 必须有真实有效的 `WHERE`；带 `JOIN` 的更新或删除必须写 `ON`；单条 `INSERT` / `REPLACE ... VALUES` 最多 1000 行。 |
| 影响行数 | `UPDATE`、`DELETE` 会执行 `EXPLAIN`，当前按执行计划首行估算，影响或扫描行数不得超过 100000。 |

明确禁止：

- `CREATE TABLE ... AS SELECT`
- `CREATE TABLE ... LIKE ...`
- 外键
- `DROP PRIMARY KEY`，以及直接删除主键列

## 3. 当前允许，但需谨慎使用

- 允许分区表、视图和跨库对象引用。
- 允许 `REPLACE`、`INSERT ... SELECT`、`ON DUPLICATE KEY UPDATE`。
- 允许 DML 使用子查询、`ORDER BY` 和 `LIMIT`。
- 允许修改列类型、使用 `CHANGE COLUMN` 改列名、重命名索引和表。
- 允许删除普通列、二级索引和表，也允许 `TRUNCATE TABLE`。
- MySQL 同一张表的多条 `ALTER` 应合并；TiDB 是否可合并需结合实际版本判断。

`DROP TABLE`、`TRUNCATE TABLE`、重命名、跨库操作和列类型变更虽然能通过当前配置，但都应经过数据影响评估、备份或回滚方案确认。

## 4. 团队建议基线

当前配置未强制表/列注释、审计时间列、列 `NOT NULL`、索引名前缀和冗余索引检查。开发仍建议：

- 默认使用 `utf8mb4`，索引采用 `idx_`、`uniq_`、`full_` 前缀。
- `INSERT` / `REPLACE` 始终显式写列名。
- 非必要不使用 `FLOAT` / `DOUBLE` 存储精确数值，优先使用合适的整数或 `DECIMAL(p,s)`。
- 不创建重复或前缀冗余索引，例如同时创建 `(a)` 和 `(a,b)` 前先确认查询收益。
- DML 使用主键或高选择性索引作为条件，禁止用 `WHERE 1=1` 等恒真条件规避检查。

参考建表：

```sql
CREATE TABLE example (
  id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '主键',
  name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '名称',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
    ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_name (name)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_general_ci
  COMMENT='示例表';
```

## 5. 审核边界与已知注意项

- 当前 `WHERE` 检查只判断语法上是否存在条件，无法判断业务条件是否安全。
- 当前配置允许冗余索引，审核器不会阻止所有“不同名称但字段重复/前缀重叠”的索引。
- 配置中的 `VARCHAR` 长度上限为 65536，高于 MySQL 常见单列声明边界；开发侧按不超过 65535 处理，并同时考虑字符集、索引长度和 65535 字节行大小限制。
- 请求级自定义参数可以临时覆盖部分规则，因此最终以工单实际返回结果和目标数据库状态为准。
