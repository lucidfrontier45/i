
## CodeGraph

This project has a CodeGraph index — a tree-sitter-parsed knowledge graph of every symbol, edge, and file.

Use CodeGraph first for structural questions: definitions, signatures, callers/callees, impact, architecture, data flow, file structure, and call paths.
Prefer `codegraph_explore` for broad questions and flows.
Use `codegraph_search` for symbol lookup, `codegraph_node` for a known symbol, `codegraph_callers` and `codegraph_callees` for relationships, `codegraph_impact` for blast radius, `codegraph_files` for structure, and `codegraph_status` for index health.
Use native search/read only for literal strings, comments, logs, config text, or a narrow range already identified by CodeGraph.
