#!/usr/bin/env python3
"""Patch provider-code-spec.json after tfplugingen-openapi.

The pfSense object `id` comes from the read operation's required query
parameter, so the generator marks it computed_optional. It is assigned by
the API on create (a config array index) and must never be user-set —
flip it to pure computed.
"""

import json
from pathlib import Path

SPEC = Path(__file__).resolve().parent.parent / "provider-code-spec.json"


def main():
    spec = json.loads(SPEC.read_text())
    for resource in spec.get("resources", []):
        for attr in resource["schema"]["attributes"]:
            if attr["name"] != "id":
                continue
            for kind, body in attr.items():
                if isinstance(body, dict) and "computed_optional_required" in body:
                    body["computed_optional_required"] = "computed"
                    print(f"{resource['name']}.id ({kind}) -> computed")
    SPEC.write_text(json.dumps(spec, indent=2, ensure_ascii=False) + "\n")


if __name__ == "__main__":
    main()
