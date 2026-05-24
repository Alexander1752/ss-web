#!/usr/bin/env python3
"""
Computes a per-member Git contribution table (lines added, removed, commits).
Merges multiple email aliases into a single canonical name.
Usage: python3 scripts/git_contributions.sh
   or: ./scripts/git_contributions.sh
"""

import subprocess
from collections import defaultdict

# ---------------------------------------------------------------------------
# Email-to-canonical-name mapping.
# Add or adjust entries here if team members use additional aliases.
# ---------------------------------------------------------------------------
EMAIL_TO_NAME = {

}

added   = defaultdict(int)
removed = defaultdict(int)
commits = defaultdict(int)

log = subprocess.check_output(
    ["git", "log", "--numstat", "--pretty=format:AUTHOR:%ae"],
    text=True,
    encoding="utf-8",
)

current = None
for line in log.splitlines():
    if line.startswith("AUTHOR:"):
        email = line[len("AUTHOR:"):].strip().lower()
        current = EMAIL_TO_NAME.get(email)  # None → skip unmapped authors
        if current:
            commits[current] += 1
    elif current and line and line[0].isdigit():
        parts = line.split("\t")
        if len(parts) >= 2 and parts[0].isdigit() and parts[1].isdigit():
            added[current]   += int(parts[0])
            removed[current] += int(parts[1])

# ---------------------------------------------------------------------------
# Print markdown table, sorted by commits descending
# ---------------------------------------------------------------------------
members = sorted(commits.keys(), key=lambda n: commits[n], reverse=True)

name_w = max(len("Team Member"), max(len(n) for n in members))
print(f"\n| {'Team Member':<{name_w}} | {'Lines Added':>11} | {'Lines Removed':>13} | {'Commits':>7} |")
print(f"|{'-'*(name_w+2)}|{'-'*13}|{'-'*15}|{'-'*9}|")
for name in members:
    print(f"| {name:<{name_w}} | {added[name]:>11,} | {removed[name]:>13,} | {commits[name]:>7} |")
print()
