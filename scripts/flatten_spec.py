#!/usr/bin/env python3
"""Flatten allOf compositions in the pfSense OpenAPI spec.

tfplugingen-openapi (tech preview) does not support schema composition:
  "found 2 allOf subschema(s), schema composition is currently not supported"

The pfSense REST API v2 spec composes every request/response schema with
allOf — e.g. responses are `allOf [Success, {data: <Model>}]` and create
bodies are `allOf [<Model>, {required: [...]}]`. This script resolves
component $refs inside allOf lists and merges the subschemas into a single
object schema (later subschemas override earlier ones per property key,
`required` lists are unioned).

Reads  spec/openapi.json
Writes spec/openapi.flattened.json  (input to tfplugingen-openapi)

Also simplifies `oneOf: [integer, string]` parameter schemas (the pfSense
object `id`) to plain integer, since the generator does not support oneOf
either and live ids are integers (config array indices).
"""

import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
SRC = ROOT / "spec" / "openapi.json"
DST = ROOT / "spec" / "openapi.flattened.json"


def resolve(node, components):
    """Return the schema a node refers to (follows one level of $ref)."""
    if isinstance(node, dict) and "$ref" in node:
        name = node["$ref"].rsplit("/", 1)[-1]
        return components[name]
    return node


def merge_allof(schemas, components):
    """Merge a list of (resolved) allOf subschemas into one object schema."""
    out = {}
    for sub in schemas:
        sub = resolve(sub, components)
        sub = flatten(sub, components)
        for key, value in sub.items():
            if key == "properties":
                props = out.setdefault("properties", {})
                props.update(value)  # later subschema wins per property
            elif key == "required":
                merged = out.get("required", []) + value
                out["required"] = sorted(set(merged))
            else:
                out[key] = value
    out.setdefault("type", "object")
    return out


def flatten(node, components):
    """Recursively flatten allOf and simplify oneOf[int,str] scalars."""
    if isinstance(node, list):
        return [flatten(item, components) for item in node]
    if not isinstance(node, dict):
        return node
    if "allOf" in node:
        rest = {k: v for k, v in node.items() if k != "allOf"}
        merged = merge_allof(node["allOf"], components)
        merged.update({k: flatten(v, components) for k, v in rest.items()})
        return merged
    # oneOf [integer, string] (the pfSense `id`) -> integer
    one_of = node.get("oneOf")
    if (
        one_of
        and len(one_of) == 2
        and {sub.get("type") for sub in one_of} == {"integer", "string"}
    ):
        rest = {k: v for k, v in node.items() if k != "oneOf"}
        rest["type"] = "integer"
        return rest
    return {k: flatten(v, components) for k, v in node.items()}


def main():
    spec = json.loads(SRC.read_text())
    components = spec.get("components", {}).get("schemas", {})
    spec["paths"] = flatten(spec["paths"], components)
    DST.write_text(json.dumps(spec, indent=2, ensure_ascii=False) + "\n")
    print(f"wrote {DST.relative_to(ROOT)}")


if __name__ == "__main__":
    main()
