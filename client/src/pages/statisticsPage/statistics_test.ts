import { describe, expect, it } from 'vitest';

import { getAvizStats, getControlStats, type StatisticsPhoto } from './statistics';

const asRecord = (rows: { name: string; value: number }[]) =>
  Object.fromEntries(rows.map((row) => [row.name, row.value]));

describe('statistics calculations', () => {
  it('counts control types independently for each photo', () => {
    const photos: StatisticsPhoto[] = [
      { control_angajare: true, control_periodic: true },
      { control_periodic: true, control_adaptare: true },
      { control_reluare: true, control_supraveghere: true, control_alte: true },
      {},
    ];

    expect(asRecord(getControlStats(photos))).toEqual({
      Angajare: 1,
      Periodic: 2,
      Adaptare: 1,
      Reluare: 1,
      Supraveghere: 1,
      Alte: 1,
    });
  });

  it('counts medical opinion fields independently for each photo', () => {
    const photos: StatisticsPhoto[] = [
      { aviz_apt: true },
      { aviz_apt: true, aviz_apt_conditionat: true },
      { aviz_inapt_temporar: true },
      { aviz_inapt: true },
      {},
    ];

    expect(asRecord(getAvizStats(photos))).toEqual({
      APT: 2,
      'APT Conditionat': 1,
      'Inapt Temporar': 1,
      Inapt: 1,
    });
  });

  it('returns zero counts when no photos are available', () => {
    expect(asRecord(getControlStats([]))).toEqual({
      Angajare: 0,
      Periodic: 0,
      Adaptare: 0,
      Reluare: 0,
      Supraveghere: 0,
      Alte: 0,
    });
    expect(asRecord(getAvizStats([]))).toEqual({
      APT: 0,
      'APT Conditionat': 0,
      'Inapt Temporar': 0,
      Inapt: 0,
    });
  });
});
