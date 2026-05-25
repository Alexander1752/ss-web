#!/usr/bin/env python3
"""
Generates the medical-data reports required for section 6 of the project
evaluation document. Connects to the MongoDB instance used by ss-web and
prints six aggregated reports as Markdown so the output can be pasted
directly into FirstForce_ProjectEvaluation.md.

Run AFTER seeding data with `python3 scripts/seed_data.py`.

Usage:
    pip install pymongo
    python3 scripts/medical_reports.py
"""

from datetime import datetime, timedelta, timezone

import pymongo
import os

MONGO_URI = f"mongodb://admin:supersecret@{os.environ.get('MONGO_HOST', '127.0.0.1')}:{os.environ.get('MONGO_PORT', '27015')}/?authSource=admin"
DB_NAME = "mqtt-streaming-server"
COLLECTION_NAME = "photos"


def md_table(headers, rows):
    widths = [max(len(str(h)), *(len(str(r[i])) for r in rows)) for i, h in enumerate(headers)]
    out = ["| " + " | ".join(str(h).ljust(widths[i]) for i, h in enumerate(headers)) + " |"]
    out.append("|" + "|".join("-" * (w + 2) for w in widths) + "|")
    for r in rows:
        out.append("| " + " | ".join(str(v).ljust(widths[i]) for i, v in enumerate(r)) + " |")
    return "\n".join(out)


def main():
    client = pymongo.MongoClient(MONGO_URI)
    coll = client[DB_NAME][COLLECTION_NAME]

    total = coll.count_documents({})
    print(f"## Report 1 — Total medical records processed\n\n**Total documents:** {total}\n")

    print("## Report 2 — Distribution of medical opinions (Aviz Medical)\n")
    pipeline = [
        {"$group": {
            "_id": None,
            "APT": {"$sum": {"$cond": ["$aviz_apt", 1, 0]}},
            "APT_CONDITIONAT": {"$sum": {"$cond": ["$aviz_apt_conditionat", 1, 0]}},
            "INAPT_TEMPORAR": {"$sum": {"$cond": ["$aviz_inapt_temporar", 1, 0]}},
            "INAPT": {"$sum": {"$cond": ["$aviz_inapt", 1, 0]}},
        }}
    ]
    res = list(coll.aggregate(pipeline))
    if res:
        r = res[0]
        rows = [
            ["APT (Fit)", r["APT"], f"{(r['APT'] / total * 100):.1f}%" if total else "—"],
            ["APT CONDITIONAT (Fit with conditions)", r["APT_CONDITIONAT"], f"{(r['APT_CONDITIONAT'] / total * 100):.1f}%" if total else "—"],
            ["INAPT TEMPORAR (Temporarily Unfit)", r["INAPT_TEMPORAR"], f"{(r['INAPT_TEMPORAR'] / total * 100):.1f}%" if total else "—"],
            ["INAPT (Unfit)", r["INAPT"], f"{(r['INAPT'] / total * 100):.1f}%" if total else "—"],
        ]
        print(md_table(["Medical Opinion", "Count", "Share"], rows))
    print()

    print("## Report 3 — Distribution of control types (Tip Control)\n")
    pipeline = [
        {"$group": {
            "_id": None,
            "Angajare": {"$sum": {"$cond": ["$control_angajare", 1, 0]}},
            "Periodic": {"$sum": {"$cond": ["$control_periodic", 1, 0]}},
            "Adaptare": {"$sum": {"$cond": ["$control_adaptare", 1, 0]}},
            "Reluare": {"$sum": {"$cond": ["$control_reluare", 1, 0]}},
            "Supraveghere": {"$sum": {"$cond": ["$control_supraveghere", 1, 0]}},
            "Alte": {"$sum": {"$cond": ["$control_alte", 1, 0]}},
        }}
    ]
    res = list(coll.aggregate(pipeline))
    if res:
        r = res[0]
        rows = [[k, r[k]] for k in ["Angajare", "Periodic", "Adaptare", "Reluare", "Supraveghere", "Alte"]]
        print(md_table(["Control Type", "Count"], rows))
    print()

    print("## Report 4 — Top 5 professions and how many of each are fit (APT)\n")
    pipeline = [
        {"$group": {
            "_id": "$profesie_functie",
            "total": {"$sum": 1},
            "fit": {"$sum": {"$cond": ["$aviz_apt", 1, 0]}},
        }},
        {"$sort": {"total": -1}},
        {"$limit": 5},
    ]
    rows = []
    for doc in coll.aggregate(pipeline):
        prof = doc["_id"] or "(unspecified)"
        pct = f"{(doc['fit'] / doc['total'] * 100):.0f}%" if doc["total"] else "—"
        rows.append([prof, doc["total"], doc["fit"], pct])
    if rows:
        print(md_table(["Profession", "Total", "Fit (APT)", "Fit %"], rows))
    print()

    print("## Report 5 — People needing a re-examination in the next 30 days\n")
    now = datetime.now(timezone.utc)
    in_30 = now + timedelta(days=30)
    cur = coll.find(
        {"data_urm_examinari": {"$gte": now, "$lte": in_30}},
        {"nume": 1, "prenume": 1, "profesie_functie": 1, "data_urm_examinari": 1},
    ).sort("data_urm_examinari", 1)
    rows = []
    for d in cur:
        rows.append([
            f"{d.get('nume', '')} {d.get('prenume', '')}".strip(),
            d.get("profesie_functie", ""),
            d.get("data_urm_examinari").strftime("%Y-%m-%d") if d.get("data_urm_examinari") else "",
        ])
    if rows:
        print(f"**Total due within 30 days:** {len(rows)}\n")
        print(md_table(["Person", "Profession", "Next Exam Date"], rows[:20]))
        if len(rows) > 20:
            print(f"\n_(showing first 20 of {len(rows)})_")
    else:
        print("_No records with `data_urm_examinari` in the next 30 days._")
    print()

    print("## Report 6 — Records per medical unit (top 5)\n")
    pipeline = [
        {"$group": {"_id": "$unitate_medicala", "count": {"$sum": 1}}},
        {"$sort": {"count": -1}},
        {"$limit": 5},
    ]
    rows = [[d["_id"] or "(unspecified)", d["count"]] for d in coll.aggregate(pipeline)]
    if rows:
        print(md_table(["Medical Unit", "Records"], rows))
    print()

    print("## Report 7 (bonus) — Fitness rate per device\n")
    pipeline = [
        {"$group": {
            "_id": "$device_id",
            "total": {"$sum": 1},
            "fit": {"$sum": {"$cond": ["$aviz_apt", 1, 0]}},
            "unfit": {"$sum": {"$cond": ["$aviz_inapt", 1, 0]}},
        }},
        {"$sort": {"total": -1}},
        {"$limit": 10},
    ]
    rows = []
    for d in coll.aggregate(pipeline):
        pct = f"{(d['fit'] / d['total'] * 100):.0f}%" if d["total"] else "—"
        rows.append([d["_id"] or "(unknown)", d["total"], d["fit"], d["unfit"], pct])
    if rows:
        print(md_table(["Device ID", "Total", "Fit", "Unfit", "Fit %"], rows))
    print()


if __name__ == "__main__":
    main()
