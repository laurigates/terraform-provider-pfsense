#!/usr/bin/env python3
"""Patch provider-code-spec.json after tfplugingen-openapi.

Two post-processing steps the generator cannot express from the OpenAPI spec:

1. The pfSense object `id` comes from the read operation's required query
   parameter, so the generator marks it computed_optional. It is assigned by
   the API on create (a config array index) and must never be user-set —
   flip it to pure computed.

2. Inject an `apply_immediately` attribute into every resource. pfSense stages
   config changes and applies them separately (POST /api/v2/.../apply). This
   attribute (default true) drives whether the shared CRUD helper applies the
   change to the running firewall after each create/update/delete. It is not in
   the API model schema — it controls client behaviour — so it is injected here
   rather than generated.
"""

import json
from pathlib import Path

SPEC = Path(__file__).resolve().parent.parent / "provider-code-spec.json"

APPLY_IMMEDIATELY_ATTR = {
    "name": "apply_immediately",
    "bool": {
        "computed_optional_required": "computed_optional",
        "default": {"static": True},
        "description": (
            "Whether to apply the staged change to the running firewall "
            "immediately after this resource is created, updated, or deleted. "
            "Defaults to true. Set to false to batch multiple changes and apply "
            "them separately."
        ),
    },
}


def main():
    spec = json.loads(SPEC.read_text())
    for resource in spec.get("resources", []):
        attrs = resource["schema"]["attributes"]

        for attr in attrs:
            if attr["name"] != "id":
                continue
            for kind, body in attr.items():
                if isinstance(body, dict) and "computed_optional_required" in body:
                    body["computed_optional_required"] = "computed"
                    print(f"{resource['name']}.id ({kind}) -> computed")

        if not any(a["name"] == "apply_immediately" for a in attrs):
            attrs.append(APPLY_IMMEDIATELY_ATTR)
            print(f"{resource['name']}.apply_immediately -> injected (default true)")

    SPEC.write_text(json.dumps(spec, indent=2, ensure_ascii=False) + "\n")


if __name__ == "__main__":
    main()
