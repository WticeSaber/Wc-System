# 2026 世界杯球队数据（唯一维护入口）

本目录中的 **`wc2026_teams.json`** 是前后端共用的球队名单数据源。  
增删改球队时**只改这一份文件**，然后重启 Go 后端；前端 dev 模式会自动热更新。

> `fm_export.csv` 等用户私有文件仍放在 `data/` 下，已被 `.gitignore` 忽略；  
> `wc2026_teams.json` 与 `README.md` 会纳入版本控制。

---

## 文件位置

| 文件 | 用途 |
|------|------|
| `data/wc2026_teams.json` | 48 支参赛队元数据（**请在此维护**） |
| `data/README.md` | 字段说明与维护指南（本文件） |

---

## 谁在读这个文件？

| 模块 | 读取方式 |
|------|----------|
| Go 后端 `internal/teams` | 启动时加载；可通过 `TEAMS_DATA_PATH` 覆盖路径 |
| Go `/api/teams` | 合并 JSON 元数据 + 实时 Elo/进球统计 |
| 前端 `frontend/src/data/teamsRegistry.ts` | 直接 import 同一份 JSON（离线降级） |
| 前端下拉框 | 优先 `/api/teams`，失败则用 JSON 本地列表 |

---

## JSON 顶层结构

```json
{
  "version": "2026.1",
  "tournament": "2026 FIFA World Cup",
  "updated_at": "2026-06-23",
  "teams": [ ... ]
}
```

---

## 单支球队字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `id` | string | ✓ | 三字母球队 ID，如 `ARG`、`USA` |
| `name_en` | string | ✓ | 英文显示名 |
| `name_cn` | string | ✓ | 中文显示名 |
| `flag_code` | string | ✓ | [FlagCDN](https://flagcdn.com) 代码，如 `ar`、`gb-eng` |
| `iso2` | string | ✓ | ISO 3166-1 alpha-2，供 World Bank API 使用 |
| `canonical_name` | string | ✓ | 与外部 CSV 对齐的标准英文名，用于 Elo/赛果匹配 |
| `csv_aliases` | string[] | ✓ | martj42 / JGravier CSV 中可能出现的别名 |
| `elo_code` | string | ✓ | [eloratings.net](https://www.eloratings.net) 两字母国家代码，如 `AR`、`EN`、`SQ`（苏格兰） |
| `group` | string | ✓ | 世界杯小组：`A`–`L` |
| `is_host` | boolean | ✓ | 是否为 2026 东道主（美国/墨西哥/加拿大） |
| `wiki_title` | string | 可选 | 英文维基球队条目 slug，用于 Wikimedia 适配器 |

### `canonical_name` 与 `csv_aliases` 怎么填？

外部数据源队名不统一，例如：

- CSV 里写 `United States` → `canonical_name` 为 `USA`，别名包含 `United States`
- CSV 里写 `Korea Republic` → `canonical_name` 为 `South Korea`

修改后若某队 Elo 或近 10 场数据为 0，优先检查 `canonical_name` / `csv_aliases` 是否与 CSV 一致。

---

## 维护示例：新增或修改一支球队

```json
{
  "id": "XXX",
  "name_en": "Example Country",
  "name_cn": "示例国",
  "flag_code": "xx",
  "iso2": "xx",
  "canonical_name": "Example Country",
  "csv_aliases": ["Example Country", "Example"],
  "group": "A",
  "is_host": false,
  "wiki_title": "Example_Country_national_football_team"
}
```

1. 在 `teams` 数组中加入或修改对象（保持 **48 支** 或你需要的数量）
2. 保存文件
3. 重启后端：`go run . serve`
4. 刷新浏览器

---

## 2026 世界杯 48 队分组速查

| 小组 | 球队 |
|------|------|
| A | 墨西哥、南非、韩国、捷克 |
| B | 加拿大、波黑、卡塔尔、瑞士 |
| C | 巴西、摩洛哥、海地、苏格兰 |
| D | 美国、巴拉圭、澳大利亚、土耳其 |
| E | 德国、库拉索、科特迪瓦、厄瓜多尔 |
| F | 荷兰、日本、瑞典、突尼斯 |
| G | 比利时、埃及、伊朗、新西兰 |
| H | 西班牙、佛得角、沙特、乌拉圭 |
| I | 法国、塞内加尔、伊拉克、挪威 |
| J | 阿根廷、阿尔及利亚、奥地利、约旦 |
| K | 葡萄牙、刚果（金）、乌兹别克斯坦、哥伦比亚 |
| L | 英格兰、克罗地亚、加纳、巴拿马 |

---

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `TEAMS_DATA_PATH` | `data/wc2026_teams.json` | 自定义球队 JSON 路径 |

Docker 镜像已包含 `data/wc2026_teams.json`；挂载自定义文件时可映射到容器内同路径。
